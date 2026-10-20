// This file runs the concurrent simulation (go run . -simulate): every
// patient is a goroutine that falls asleep and wakes up at random times,
// and one more goroutine keeps moving hallway patients into free rooms.
// All of them share one Hospital, which protects itself with a mutex.

package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"JuegoHospitalGo/hospital"
)

// locations are the places where a simulated attack can happen.
var locations = []string{"Cafeteria", "Waiting room", "Pharmacy", "Elevator"}

// maxSleep is the longest a patient may stay asleep; after that they are
// woken up automatically (see wakeLongSleepers).
const maxSleep = 30 * time.Second

// pace sets how long patients stay awake and asleep in the simulation.
type pace struct {
	awakeMin, awakeMax   time.Duration
	asleepMin, asleepMax time.Duration
}

// terminalPace is fast so that a short -simulate run shows many events;
// gamePace is slower so that the web game can be followed by eye.
var (
	terminalPace = pace{500 * time.Millisecond, 2 * time.Second, time.Second, 3 * time.Second}
	gamePace     = pace{4 * time.Second, 12 * time.Second, 10 * time.Second, 25 * time.Second}
)

// runSimulation builds the hospital of the scenario and lets the patients
// live on their own for the given duration.
func runSimulation(duration time.Duration) error {
	h, err := hospital.NewHospital("My Hospital")
	if err != nil {
		return err
	}
	c, err := setUp(h)
	if err != nil {
		return err
	}

	printHeader(fmt.Sprintf("Concurrent simulation (%s)", duration))
	start := time.Now()
	stop := make(chan struct{})
	var wg sync.WaitGroup

	for _, p := range c.patients {
		wg.Add(1)
		go func(p *hospital.Patient) {
			defer wg.Done()
			simulatePatient(h, p, terminalPace, true, start, stop)
		}(p)
	}
	wg.Add(2)
	go func() {
		defer wg.Done()
		moveHallwayPatients(h, start, stop)
	}()
	go func() {
		defer wg.Done()
		wakeLongSleepers(h, start, stop)
	}()

	time.Sleep(duration)
	close(stop) // tells every goroutine to finish
	wg.Wait()   // waits until they all have

	// Every goroutine has finished, so reading the hospital directly is safe.
	printQueries(h, c)
	return nil
}

// simulatePatient makes p fall asleep again and again, at the given pace,
// until stop is closed. If wakeAlone is true the patient also wakes up by
// themselves; if not, the goroutine waits until somebody else (the player)
// wakes them up before the next attack.
func simulatePatient(h *hospital.Hospital, p *hospital.Patient, pc pace, wakeAlone bool, start time.Time, stop <-chan struct{}) {
	for {
		if !wait(stop, pc.awakeMin, pc.awakeMax) {
			return
		}
		location := locations[rand.Intn(len(locations))]
		episode, err := h.HandleSleepAttack(p, location)
		switch {
		case errors.Is(err, hospital.ErrNoRoomAvailable):
			logf(start, "%s fell asleep at %s, attended by %s; no free room, stays in the hallway",
				p.Name(), location, episode.AttendedBy().Name())
		case err != nil:
			logf(start, "%s: %v", p.Name(), err)
			continue
		default:
			logf(start, "%s fell asleep at %s, attended by %s -> room %s",
				p.Name(), location, episode.AttendedBy().Name(), episode.Room().ID())
		}

		if !wakeAlone {
			// Manual mode: wait (checking once a second) until the player
			// wakes the patient up.
			for !isAwake(h, p.ID()) {
				if !wait(stop, time.Second, time.Second) {
					return
				}
			}
			continue
		}
		if !wait(stop, pc.asleepMin, pc.asleepMax) {
			return
		}
		if err := h.WakePatient(p); err == nil {
			logf(start, "%s woke up", p.Name())
		}
	}
}

// isAwake reports whether the patient with that id is awake. It reads the
// state through Snapshot, so it is safe while other goroutines change it.
func isAwake(h *hospital.Hospital, id string) bool {
	for _, p := range h.Snapshot().Patients {
		if p.ID == id {
			return p.State == hospital.Awake
		}
	}
	return false
}

// wakeLongSleepers checks every second which patients are asleep and wakes
// up anyone who has been asleep longer than maxSleep, until stop is closed.
// It only reads the hospital through Snapshot, so it is safe to run next to
// the other goroutines.
func wakeLongSleepers(h *hospital.Hospital, start time.Time, stop <-chan struct{}) {
	asleepSince := make(map[string]time.Time) // patient id -> when we first saw them asleep
	for wait(stop, time.Second, time.Second) {
		for _, p := range h.Snapshot().Patients {
			asleep := p.State == hospital.AsleepInHallway || p.State == hospital.AsleepInBed
			if !asleep {
				delete(asleepSince, p.ID)
				continue
			}
			since, seen := asleepSince[p.ID]
			if !seen {
				asleepSince[p.ID] = time.Now()
				continue
			}
			if time.Since(since) < maxSleep {
				continue
			}
			patient, err := h.FindPatient(p.ID)
			if err == nil && h.WakePatient(patient) == nil {
				logf(start, "%s woke up after %s asleep (maximum)", p.Name, maxSleep)
				delete(asleepSince, p.ID)
			}
		}
	}
}

// moveHallwayPatients tries every few moments to give a room to the
// patients sleeping in the hallway, until stop is closed.
func moveHallwayPatients(h *hospital.Hospital, start time.Time, stop <-chan struct{}) {
	for wait(stop, 300*time.Millisecond, 300*time.Millisecond) {
		for _, p := range h.PatientsInHallway() {
			// Another goroutine may have woken p up in the meantime; then
			// AssignRoom just returns an error and nothing changes.
			if room, err := h.AssignRoom(p); err == nil {
				logf(start, "%s was moved from the hallway to room %s", p.Name(), room.ID())
			}
		}
	}
}

// wait sleeps a random time between min and max. It returns false if stop
// was closed in the meantime.
func wait(stop <-chan struct{}, min, max time.Duration) bool {
	d := min
	if max > min {
		d += time.Duration(rand.Int63n(int64(max - min)))
	}
	select {
	case <-stop:
		return false
	case <-time.After(d):
		return true
	}
}

// logf prints one simulation event with the seconds since start.
func logf(start time.Time, format string, args ...any) {
	fmt.Printf("  [%4.1fs] %s\n", time.Since(start).Seconds(), fmt.Sprintf(format, args...))
}
