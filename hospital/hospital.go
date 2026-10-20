// Package hospital models a hospital for patients with narcolepsy: patients,
// their sleep episodes, room assignment and the staff that attends them.
//
// The package contains business logic only; it performs no printing or
// presentation, which is the responsibility of the caller.
package hospital

import (
	"fmt"
	"strings"
	"time"
)

// Hospital is the aggregate root of the package: it owns the patients, the
// staff, the rooms and the episode history, and it is the only type that
// coordinates changes spanning more than one of them (for example, moving a
// patient to a bed and occupying the room at the same time).
//
// Collections are slices rather than maps so that every query returns its
// results in a deterministic order (admission, hiring or registration order).
type Hospital struct {
	name     string
	patients []*Patient
	staff    []Attender
	rooms    []*Room
	episodes []EpisodeRecord

	episodeSeq   int              // last episode number issued; ids are E-001, E-002, ...
	nextAttender int              // index where the next round-robin dispatch starts
	now          func() time.Time // clock; replaced in tests for deterministic dates
}

// NewHospital builds an empty hospital with the given name.
//
// It returns an error wrapping ErrInvalidName when name is blank.
func NewHospital(name string) (*Hospital, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("new hospital: %w", ErrInvalidName)
	}
	return &Hospital{name: name, now: time.Now}, nil
}

// Name returns the hospital name.
func (h *Hospital) Name() string { return h.name }

// AdmitPatient adds p to the hospital.
//
// It returns ErrNilPatient when p is nil and an error wrapping ErrDuplicateID
// when a patient with the same id was already admitted.
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

// HireAttender adds a staff member of any kind to the hospital.
//
// It returns ErrNilAttender when a is nil and an error wrapping
// ErrDuplicateID when a staff member with the same id was already hired.
// A typed nil pointer (for example (*Doctor)(nil)) is not detected here; use
// HireDoctor for doctors, which rejects it.
func (h *Hospital) HireAttender(a Attender) error {
	if a == nil {
		return ErrNilAttender
	}
	if h.findAttender(a.ID()) != nil {
		return fmt.Errorf("hire %q: %w", a.ID(), ErrDuplicateID)
	}
	h.staff = append(h.staff, a)
	return nil
}

// HireDoctor adds d to the hospital staff. It delegates to HireAttender and
// returns the same errors; a nil d yields ErrNilAttender.
func (h *Hospital) HireDoctor(d *Doctor) error {
	if d == nil {
		return ErrNilAttender
	}
	return h.HireAttender(d)
}

// RegisterRoom adds r to the hospital.
//
// It returns ErrNilRoom when r is nil and an error wrapping ErrDuplicateID
// when a room with the same id was already registered.
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

// Patients returns a copy of the admitted patients in admission order.
func (h *Hospital) Patients() []*Patient {
	return append([]*Patient(nil), h.patients...)
}

// Staff returns a copy of the hired staff in hiring order.
func (h *Hospital) Staff() []Attender {
	return append([]Attender(nil), h.staff...)
}

// Rooms returns a copy of the registered rooms in registration order.
func (h *Hospital) Rooms() []*Room {
	return append([]*Room(nil), h.rooms...)
}

// Episodes returns a copy of the episode history in registration order.
func (h *Hospital) Episodes() []EpisodeRecord {
	return append([]EpisodeRecord(nil), h.episodes...)
}

// HandleSleepAttack runs the full flow of a sleep attack suffered by p at
// location: the patient falls asleep in the hallway, an available staff
// member attends them, a room is assigned if one is free and the episode is
// registered.
//
// It returns ErrNilPatient or an error wrapping ErrPatientNotAdmitted when p
// is not an admitted patient, an error wrapping the Patient.SufferSleepAttack
// error when p cannot fall asleep, and an error wrapping
// ErrNoAttenderAvailable when nobody can attend p; in those cases no episode
// is registered.
//
// When no room is free the episode is still registered, with a nil Room, and
// the returned record is accompanied by an error wrapping ErrNoRoomAvailable:
// the attack did happen and was attended, but the patient stays asleep in the
// hallway.
func (h *Hospital) HandleSleepAttack(p *Patient, location string) (EpisodeRecord, error) {
	if err := h.requireAdmitted(p); err != nil {
		return EpisodeRecord{}, err
	}
	if err := p.SufferSleepAttack(location); err != nil {
		return EpisodeRecord{}, fmt.Errorf("sleep attack of %q: %w", p.ID(), err)
	}
	record, err := h.dispatch(p, p.Location())
	if err != nil {
		return EpisodeRecord{}, fmt.Errorf("sleep attack of %q: %w", p.ID(), err)
	}
	room, roomErr := h.AssignRoom(p)
	return h.registerEpisode(record, room), roomErr
}

// AssignRoom places p, who must be asleep in the hallway, in the first
// available room in registration order and moves them to AsleepInBed.
//
// It returns ErrNilPatient or an error wrapping ErrPatientNotAdmitted when p
// is not an admitted patient, an error wrapping ErrPatientNotInHallway when p
// is in any other state, and an error wrapping ErrNoRoomAvailable when every
// room is full. On error neither the patient nor any room changes, and no
// room is created.
func (h *Hospital) AssignRoom(p *Patient) (*Room, error) {
	if err := h.requireAdmitted(p); err != nil {
		return nil, err
	}
	if p.State() != AsleepInHallway {
		return nil, fmt.Errorf("assign room to %q: %w", p.ID(), ErrPatientNotInHallway)
	}
	for _, r := range h.rooms {
		if r.State() != Available {
			continue
		}
		if err := r.occupy(p); err != nil {
			return nil, fmt.Errorf("assign room to %q: %w", p.ID(), err)
		}
		if err := p.moveToBed(); err != nil {
			_ = r.release(p) // undo occupy so the room and patient stay consistent
			return nil, fmt.Errorf("assign room to %q: %w", p.ID(), err)
		}
		return r, nil
	}
	return nil, fmt.Errorf("assign room to %q: %w", p.ID(), ErrNoRoomAvailable)
}

// WakePatient wakes p up and, if p was in a bed, releases that bed. It is the
// only way to wake a patient, so a room can never stay occupied by an awake
// patient.
//
// It returns ErrNilPatient or an error wrapping ErrPatientNotAdmitted when p
// is not an admitted patient, and an error wrapping ErrPatientAwake or
// ErrPatientDischarged when p is not asleep.
func (h *Hospital) WakePatient(p *Patient) error {
	if err := h.requireAdmitted(p); err != nil {
		return err
	}
	if err := p.wakeUp(); err != nil {
		return fmt.Errorf("wake %q: %w", p.ID(), err)
	}
	if r := h.roomOf(p); r != nil {
		if err := r.release(p); err != nil {
			return fmt.Errorf("wake %q: %w", p.ID(), err)
		}
	}
	return nil
}

// PatientsInHallway returns, in admission order, the patients currently
// asleep in the hallway. The returned slice is new on every call.
func (h *Hospital) PatientsInHallway() []*Patient {
	var result []*Patient
	for _, p := range h.patients {
		if p.State() == AsleepInHallway {
			result = append(result, p)
		}
	}
	return result
}

// SevereReport returns, for every admitted patient with Severe narcolepsy,
// the number of episodes they suffered today. Severe patients without
// episodes today appear with 0, so the report lists every severe case.
// "Today" is the calendar day of the hospital clock. The returned map is new
// on every call; map iteration order is random, so callers that print it
// should sort the keys.
func (h *Hospital) SevereReport() map[*Patient]int {
	report := make(map[*Patient]int)
	for _, p := range h.patients {
		if p.Level() == Severe {
			report[p] = 0
		}
	}
	today := h.now()
	for _, e := range h.episodes {
		if _, severe := report[e.Patient()]; severe && sameDay(e.OccurredAt(), today) {
			report[e.Patient()]++
		}
	}
	return report
}

// dispatch asks the staff, in round-robin order starting after the last
// staff member dispatched, to attend p. Unavailable staff are skipped, and a
// staff member whose Attend fails (for example an Orderly with a Severe
// patient) is passed over in favor of the next one. Round-robin spreads the
// work instead of sending every patient to the first staff member hired.
//
// It returns an error wrapping ErrNoAttenderAvailable when nobody attends p.
func (h *Hospital) dispatch(p *Patient, location string) (EpisodeRecord, error) {
	n := len(h.staff)
	for i := 0; i < n; i++ {
		idx := (h.nextAttender + i) % n
		a := h.staff[idx]
		if !a.IsAvailable() {
			continue
		}
		record, err := a.Attend(p, location)
		if err != nil {
			continue
		}
		h.nextAttender = (idx + 1) % n
		return record, nil
	}
	return EpisodeRecord{}, ErrNoAttenderAvailable
}

// registerEpisode completes a record returned by Attend with the next
// episode id, the current time of the hospital clock and the assigned room
// (nil when the patient stayed in the hallway), stores it in the hospital
// history and in the history of the staff member who attended it, and
// returns the completed record.
func (h *Hospital) registerEpisode(e EpisodeRecord, room *Room) EpisodeRecord {
	h.episodeSeq++
	e.id = fmt.Sprintf("E-%03d", h.episodeSeq)
	e.occurredAt = h.now()
	e.room = room
	h.episodes = append(h.episodes, e)
	e.attendedBy.recordEpisode(e)
	return e
}

// requireAdmitted returns ErrNilPatient when p is nil and an error wrapping
// ErrPatientNotAdmitted when p is not this hospital's patient. Patients are
// compared by pointer identity, so a different *Patient with the same id is
// not considered admitted.
func (h *Hospital) requireAdmitted(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	if h.findPatient(p.ID()) != p {
		return fmt.Errorf("patient %q: %w", p.ID(), ErrPatientNotAdmitted)
	}
	return nil
}

// findPatient returns the admitted patient with the given id, or nil.
func (h *Hospital) findPatient(id string) *Patient {
	for _, p := range h.patients {
		if p.ID() == id {
			return p
		}
	}
	return nil
}

// findAttender returns the staff member with the given id, or nil.
func (h *Hospital) findAttender(id string) Attender {
	for _, a := range h.staff {
		if a.ID() == id {
			return a
		}
	}
	return nil
}

// roomOf returns the room occupied by p, or nil when p is not in a room.
func (h *Hospital) roomOf(p *Patient) *Room {
	for _, r := range h.rooms {
		if r.contains(p) {
			return r
		}
	}
	return nil
}

// sameDay reports whether a and b fall on the same calendar day in the time
// zone of b.
func sameDay(a, b time.Time) bool {
	a = a.In(b.Location())
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
