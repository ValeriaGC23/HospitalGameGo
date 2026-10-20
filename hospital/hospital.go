// Package hospital models a hospital for patients with narcolepsy: patients,
// their sleep episodes, the rooms and the staff that attends them.
//
// The package only contains the logic; it never prints anything.
package hospital

import (
	"fmt"
	"strings"
	"time"
)

// Hospital owns the patients, staff, rooms and episodes. Anything that
// changes two of them together (for example a patient and a room) goes
// through a Hospital method, so they always stay consistent.
//
// It uses slices instead of maps so that results always come out in the
// same order.
type Hospital struct {
	name     string
	patients []*Patient
	staff    []Attender
	rooms    []*Room
	episodes []EpisodeRecord

	episodeCount int              // used to build ids E-001, E-002, ...
	nextAttender int              // whose turn it is in the round-robin
	now          func() time.Time // the clock; tests replace it with a fixed date
}

// NewHospital builds an empty hospital. It returns an error if name is empty.
func NewHospital(name string) (*Hospital, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("new hospital: %w", ErrInvalidName)
	}
	return &Hospital{name: name, now: time.Now}, nil
}

// Name returns the hospital name.
func (h *Hospital) Name() string { return h.name }

// AdmitPatient adds p to the hospital. It returns ErrNilPatient or
// ErrDuplicateID.
func (h *Hospital) AdmitPatient(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	if h.findPatient(p.ID()) != nil {
		return fmt.Errorf("admit patient %q: %w", p.ID(), ErrDuplicateID)
	}
	h.patients = append(h.patients, p)
	return nil
}

// HireAttender adds any kind of staff member. It returns ErrNilAttender or
// ErrDuplicateID.
func (h *Hospital) HireAttender(a Attender) error {
	if a == nil {
		return ErrNilAttender
	}
	for _, s := range h.staff {
		if s.ID() == a.ID() {
			return fmt.Errorf("hire %q: %w", a.ID(), ErrDuplicateID)
		}
	}
	h.staff = append(h.staff, a)
	return nil
}

// HireDoctor adds a doctor. It uses HireAttender, because a *Doctor is an
// Attender.
func (h *Hospital) HireDoctor(d *Doctor) error {
	if d == nil {
		return ErrNilAttender
	}
	return h.HireAttender(d)
}

// RegisterRoom adds r to the hospital. It returns ErrNilRoom or
// ErrDuplicateID.
func (h *Hospital) RegisterRoom(r *Room) error {
	if r == nil {
		return ErrNilRoom
	}
	for _, existing := range h.rooms {
		if existing.ID() == r.ID() {
			return fmt.Errorf("register room %q: %w", r.ID(), ErrDuplicateID)
		}
	}
	h.rooms = append(h.rooms, r)
	return nil
}

// Patients returns a copy of the admitted patients.
func (h *Hospital) Patients() []*Patient { return append([]*Patient(nil), h.patients...) }

// Staff returns a copy of the hired staff.
func (h *Hospital) Staff() []Attender { return append([]Attender(nil), h.staff...) }

// Rooms returns a copy of the registered rooms.
func (h *Hospital) Rooms() []*Room { return append([]*Room(nil), h.rooms...) }

// Episodes returns a copy of every registered episode.
func (h *Hospital) Episodes() []EpisodeRecord { return append([]EpisodeRecord(nil), h.episodes...) }

// HandleSleepAttack runs the whole sleep attack of p at location:
//
//  1. the patient falls asleep in the hallway,
//  2. the next available staff member attends them,
//  3. the hospital looks for a free room,
//  4. the episode is saved.
//
// If no room is free, the episode is still saved (the attack did happen)
// and the error ErrNoRoomAvailable is returned together with the record.
// For any other error nothing is saved.
func (h *Hospital) HandleSleepAttack(p *Patient, location string) (EpisodeRecord, error) {
	if err := h.checkAdmitted(p); err != nil {
		return EpisodeRecord{}, err
	}
	if err := p.SufferSleepAttack(location); err != nil {
		return EpisodeRecord{}, fmt.Errorf("sleep attack of %q: %w", p.ID(), err)
	}
	record, err := h.dispatch(p)
	if err != nil {
		return EpisodeRecord{}, fmt.Errorf("sleep attack of %q: %w", p.ID(), err)
	}
	room, roomErr := h.AssignRoom(p)
	return h.saveEpisode(record, room), roomErr
}

// AssignRoom puts p, who must be asleep in the hallway, in the first free
// room. It returns ErrNoRoomAvailable when every room is full; in that case
// nothing changes and no room is created.
func (h *Hospital) AssignRoom(p *Patient) (*Room, error) {
	if err := h.checkAdmitted(p); err != nil {
		return nil, err
	}
	if p.State() != AsleepInHallway {
		return nil, fmt.Errorf("assign room to %q: %w", p.ID(), ErrPatientNotInHallway)
	}
	for _, r := range h.rooms {
		if r.State() == Available {
			r.occupy(p)
			p.moveToBed()
			return r, nil
		}
	}
	return nil, fmt.Errorf("assign room to %q: %w", p.ID(), ErrNoRoomAvailable)
}

// WakePatient wakes p up and frees their bed. Waking up and freeing the bed
// happen together here, so a room never keeps an awake patient.
func (h *Hospital) WakePatient(p *Patient) error {
	if err := h.checkAdmitted(p); err != nil {
		return err
	}
	if err := p.wakeUp(); err != nil {
		return fmt.Errorf("wake %q: %w", p.ID(), err)
	}
	for _, r := range h.rooms {
		if r.contains(p) {
			return r.release(p)
		}
	}
	return nil
}

// PatientsInHallway returns the patients asleep in the hallway.
func (h *Hospital) PatientsInHallway() []*Patient {
	var result []*Patient
	for _, p := range h.patients {
		if p.State() == AsleepInHallway {
			result = append(result, p)
		}
	}
	return result
}

// SevereReport returns how many episodes each Severe patient had today.
// Severe patients with no episodes today appear with 0.
func (h *Hospital) SevereReport() map[*Patient]int {
	report := make(map[*Patient]int)
	for _, p := range h.patients {
		if p.Level() == Severe {
			report[p] = 0
		}
	}
	today := h.now()
	for _, e := range h.episodes {
		if _, isSevere := report[e.Patient()]; isSevere && sameDay(e.OccurredAt(), today) {
			report[e.Patient()]++
		}
	}
	return report
}

// dispatch finds a staff member to attend p, taking turns (round-robin) so
// that the work is shared. Staff who are unavailable, or who refuse the
// patient (an Orderly with a Severe patient), are skipped.
func (h *Hospital) dispatch(p *Patient) (EpisodeRecord, error) {
	for i := 0; i < len(h.staff); i++ {
		turn := (h.nextAttender + i) % len(h.staff)
		a := h.staff[turn]
		if !a.IsAvailable() {
			continue
		}
		record, err := a.Attend(p, p.Location())
		if err == nil {
			h.nextAttender = turn + 1
			return record, nil
		}
	}
	return EpisodeRecord{}, ErrNoAttenderAvailable
}

// saveEpisode adds the id, the time and the room to the record, and saves it
// in the hospital history and in the history of the staff member.
func (h *Hospital) saveEpisode(e EpisodeRecord, room *Room) EpisodeRecord {
	h.episodeCount++
	e.id = fmt.Sprintf("E-%03d", h.episodeCount)
	e.occurredAt = h.now()
	e.room = room
	h.episodes = append(h.episodes, e)
	e.attendedBy.recordEpisode(e)
	return e
}

// checkAdmitted returns an error if p is nil or is not a patient of this
// hospital.
func (h *Hospital) checkAdmitted(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	if h.findPatient(p.ID()) != p {
		return fmt.Errorf("patient %q: %w", p.ID(), ErrPatientNotAdmitted)
	}
	return nil
}

// findPatient returns the patient with that id, or nil.
func (h *Hospital) findPatient(id string) *Patient {
	for _, p := range h.patients {
		if p.ID() == id {
			return p
		}
	}
	return nil
}

// sameDay reports whether a and b are on the same calendar day.
func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}
