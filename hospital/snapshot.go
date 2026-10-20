// This file defines Snapshot, a copy of the whole hospital taken while
// holding the lock. Programs that run goroutines (the simulation, the web
// interface) read the hospital only through a Snapshot, so they never read
// a patient or room while another goroutine is changing it.
//
// The snapshot types are plain data with exported fields on purpose: they
// are copies, so changing them cannot affect the hospital.

package hospital

import "time"

// Snapshot is the state of the hospital at one moment.
type Snapshot struct {
	Name        string
	Patients    []PatientSnapshot
	Staff       []StaffSnapshot
	Rooms       []RoomSnapshot
	Episodes    []EpisodeSnapshot
	SevereToday map[string]int // patient id -> episodes today (SevereReport)
}

// PatientSnapshot is a copy of one patient.
type PatientSnapshot struct {
	ID       string
	Name     string
	Age      int
	Level    NarcolepsyLevel
	State    PatientState
	Location string
	RoomID   string // empty if the patient is not in a room
}

// StaffSnapshot is a copy of one staff member.
type StaffSnapshot struct {
	ID        string
	Name      string
	Role      string
	Available bool
	Episodes  int
}

// RoomSnapshot is a copy of one room.
type RoomSnapshot struct {
	ID        string
	Capacity  int
	State     RoomState
	Occupants []string // patient ids
}

// EpisodeSnapshot is a copy of one episode.
type EpisodeSnapshot struct {
	ID          string
	PatientID   string
	PatientName string
	Location    string
	AttendedBy  string
	RoomID      string // empty if the patient stayed in the hallway
	OccurredAt  time.Time
}

// Snapshot returns a copy of the current state of the hospital.
func (h *Hospital) Snapshot() Snapshot {
	h.mu.Lock()
	defer h.mu.Unlock()

	s := Snapshot{Name: h.name, SevereToday: make(map[string]int)}
	for _, p := range h.patients {
		ps := PatientSnapshot{
			ID: p.ID(), Name: p.Name(), Age: p.Age(),
			Level: p.Level(), State: p.State(), Location: p.Location(),
		}
		if r := h.roomOf(p); r != nil {
			ps.RoomID = r.ID()
		}
		s.Patients = append(s.Patients, ps)
	}
	for _, a := range h.staff {
		s.Staff = append(s.Staff, StaffSnapshot{
			ID: a.ID(), Name: a.Name(), Role: a.Role(),
			Available: a.IsAvailable(), Episodes: h.countEpisodes(a),
		})
	}
	for _, r := range h.rooms {
		rs := RoomSnapshot{ID: r.ID(), Capacity: r.Capacity(), State: r.State()}
		for _, p := range r.occupants {
			rs.Occupants = append(rs.Occupants, p.ID())
		}
		s.Rooms = append(s.Rooms, rs)
	}
	for _, e := range h.episodes {
		es := EpisodeSnapshot{
			ID: e.ID(), PatientID: e.Patient().ID(), PatientName: e.Patient().Name(),
			Location: e.Location(), AttendedBy: e.AttendedBy().Name(), OccurredAt: e.OccurredAt(),
		}
		if e.Room() != nil {
			es.RoomID = e.Room().ID()
		}
		s.Episodes = append(s.Episodes, es)
	}
	for p, n := range h.severeReport() {
		s.SevereToday[p.ID()] = n
	}
	return s
}

// countEpisodes returns how many episodes a attended. The caller holds the
// lock.
func (h *Hospital) countEpisodes(a Attender) int {
	n := 0
	for _, e := range h.episodes {
		if e.AttendedBy() == a {
			n++
		}
	}
	return n
}
