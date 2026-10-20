// Command JuegoHospitalGo runs the demonstration scenario of the narcolepsy
// hospital: it builds the hospital, triggers the sleep attacks and prints the
// result of the four queries. All business logic lives in package hospital;
// this file only builds the scenario and formats the output.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"JuegoHospitalGo/hospital"
)

// episodeKeeper is the read-only view main needs to list the episodes of a
// staff member. It is declared here, by the consumer, so that the hospital
// package does not need an extra interface for presentation purposes.
type episodeKeeper interface {
	Name() string
	MyEpisodes() []hospital.EpisodeRecord
}

// cast holds the entities created for the scenario so that later steps can
// refer to them.
type cast struct {
	doctors  []*hospital.Doctor
	orderly  *hospital.Orderly
	patients []*hospital.Patient
}

// main runs the demonstration scenario by default. With -simulate it runs
// the concurrent simulation and with -web the web interface, so "go run ."
// never changes.
func main() {
	simulate := flag.Bool("simulate", false, "run the concurrent simulation instead of the scenario")
	duration := flag.Duration("duration", 10*time.Second, "how long the simulation runs")
	web := flag.Bool("web", false, "start the web interface")
	addr := flag.String("addr", ":8080", "address of the web interface")
	flag.Parse()

	var err error
	switch {
	case *web:
		err = runWeb(*addr)
	case *simulate:
		err = runSimulation(*duration)
	default:
		err = run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run executes the four steps of the scenario in the required order. Any
// error other than the expected "no room available" aborts the demo, since
// it means the scenario itself is wrong.
func run() error {
	h, err := hospital.NewHospital("My Hospital")
	if err != nil {
		return err
	}
	c, err := setUp(h)
	if err != nil {
		return err
	}
	if err := sleepAttacks(h, c); err != nil {
		return err
	}
	if err := wakeUpAndReassign(h, c); err != nil {
		return err
	}
	printQueries(h, c)
	return nil
}

// patientData is the data needed to create one patient.
type patientData struct {
	id, name string
	age      int
	level    hospital.NarcolepsyLevel
}

// demoPatients are the 5 patients of the scenario, 2 of them Severe.
// Order matters: dispatch is round-robin (doctor, doctor, orderly, ...), so
// the third patient to fall asleep is Mild and can be attended by the
// orderly.
var demoPatients = []patientData{
	{"P-001", "Amara Okafor", 28, hospital.Severe},
	{"P-002", "Liam O'Connor", 41, hospital.Moderate},
	{"P-003", "Sofia Lindqvist", 19, hospital.Mild},
	{"P-004", "Mateo Fernandez", 63, hospital.Severe},
	{"P-005", "Priya Sharma", 35, hospital.Mild},
}

// setUp performs step 1: hires 2 doctors and 1 orderly, registers 3 rooms of
// capacity 1 and admits the 5 demo patients.
func setUp(h *hospital.Hospital) (*cast, error) {
	c := &cast{}
	if err := hireStaffAndRooms(h, c); err != nil {
		return nil, err
	}
	if err := admitPatients(h, c, demoPatients); err != nil {
		return nil, err
	}
	fmt.Printf("%s: %d staff, %d rooms, %d patients admitted\n",
		h.Name(), len(h.Staff()), len(h.Rooms()), len(h.Patients()))
	return c, nil
}

// hireStaffAndRooms hires 2 doctors and 1 orderly and registers 3 rooms of
// capacity 1.
func hireStaffAndRooms(h *hospital.Hospital, c *cast) error {
	for _, d := range []struct {
		id, name string
		age      int
	}{
		{"D-001", "Dr. Elena Rossi", 52},
		{"D-002", "Dr. Kenji Tanaka", 47},
	} {
		doctor, err := hospital.NewDoctor(d.id, d.name, d.age)
		if err != nil {
			return err
		}
		if err := h.HireDoctor(doctor); err != nil {
			return err
		}
		c.doctors = append(c.doctors, doctor)
	}

	orderly, err := hospital.NewOrderly("O-001", "Lucas Moreau", 35)
	if err != nil {
		return err
	}
	if err := h.HireAttender(orderly); err != nil {
		return err
	}
	c.orderly = orderly

	for _, id := range []string{"R-101", "R-102", "R-103"} {
		room, err := hospital.NewRoom(id, 1)
		if err != nil {
			return err
		}
		if err := h.RegisterRoom(room); err != nil {
			return err
		}
	}
	return nil
}

// admitPatients creates and admits one patient per entry of list.
func admitPatients(h *hospital.Hospital, c *cast, list []patientData) error {
	for _, p := range list {
		patient, err := hospital.NewPatient(p.id, p.name, p.age, p.level)
		if err != nil {
			return err
		}
		if err := h.AdmitPatient(patient); err != nil {
			return err
		}
		c.patients = append(c.patients, patient)
	}
	return nil
}

// sleepAttacks performs step 2: four patients fall asleep in different
// places. The first three get a room; the fourth finds every room full and
// the error is printed instead of stopping the program.
func sleepAttacks(h *hospital.Hospital, c *cast) error {
	printHeader("Sleep attacks")
	attacks := []struct {
		patient  *hospital.Patient
		location string
	}{
		{c.patients[0], "Cafeteria"},
		{c.patients[1], "Waiting room"},
		{c.patients[2], "Pharmacy"},
		{c.patients[3], "Elevator"},
	}
	for _, a := range attacks {
		episode, err := h.HandleSleepAttack(a.patient, a.location)
		switch {
		case errors.Is(err, hospital.ErrNoRoomAvailable):
			printEpisode(episode)
			fmt.Printf("  ! %s could not be given a room: %v\n", a.patient.Name(), err)
		case err != nil:
			return err
		default:
			printEpisode(episode)
		}
	}
	return nil
}

// wakeUpAndReassign performs step 3: one patient wakes up and frees a bed,
// and every patient still asleep in the hallway goes through the normal
// AssignRoom flow; nothing about the result is hard-coded.
func wakeUpAndReassign(h *hospital.Hospital, c *cast) error {
	printHeader("Room assignment")
	waking := c.patients[1]
	if err := h.WakePatient(waking); err != nil {
		return err
	}
	fmt.Printf("  %s woke up and left their bed\n", waking.Name())

	for _, p := range h.PatientsInHallway() {
		room, err := h.AssignRoom(p)
		if err != nil {
			fmt.Printf("  ! %s could not be given a room: %v\n", p.Name(), err)
			continue
		}
		fmt.Printf("  %s was moved from the hallway to room %s\n", p.Name(), room.ID())
	}
	return nil
}

// printQueries performs step 4: prints the results of the queries and the
// final state of the rooms.
func printQueries(h *hospital.Hospital, c *cast) {
	printHeader("Patients asleep in hallway")
	hallway := h.PatientsInHallway()
	if len(hallway) == 0 {
		fmt.Println("  (none)")
	}
	for _, p := range hallway {
		fmt.Printf("  %s %s (%s) at %s\n", p.ID(), p.Name(), p.Level(), p.Location())
	}

	printHeader("Episodes by staff member")
	keepers := []episodeKeeper{c.doctors[0], c.doctors[1], c.orderly}
	for _, k := range keepers {
		episodes := k.MyEpisodes()
		fmt.Printf("  %s: %d episode(s)\n", k.Name(), len(episodes))
		for _, e := range episodes {
			printEpisode(e)
		}
	}

	printHeader("Severe report (episodes today)")
	report := h.SevereReport()
	severe := make([]*hospital.Patient, 0, len(report))
	for p := range report {
		severe = append(severe, p)
	}
	// Map iteration order is random; sort for a stable output.
	sort.Slice(severe, func(i, j int) bool { return severe[i].ID() < severe[j].ID() })
	for _, p := range severe {
		fmt.Printf("  %s %-22s %d\n", p.ID(), p.Name(), report[p])
	}

	printHeader("Rooms")
	for _, r := range h.Rooms() {
		names := make([]string, 0, r.Capacity())
		for _, p := range r.Occupants() {
			names = append(names, p.Name())
		}
		fmt.Printf("  %s %-9s %s\n", r.ID(), r.State(), strings.Join(names, ", "))
	}
}

// printHeader prints a section heading in the "== Title ==" format.
func printHeader(title string) {
	fmt.Printf("\n== %s ==\n", title)
}

// printEpisode prints one episode on a single line.
func printEpisode(e hospital.EpisodeRecord) {
	room := "no room (stayed in hallway)"
	if e.Room() != nil {
		room = "room " + e.Room().ID()
	}
	p := e.Patient()
	fmt.Printf("  %s %-22s %-8s at %-26s attended by %-20s -> %s\n",
		e.ID(), p.Name(), p.Level(), e.Location(), e.AttendedBy().Name(), room)
}
