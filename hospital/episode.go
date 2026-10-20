// This file defines EpisodeRecord, the record of one sleep attack. It is a
// value (not a pointer) and its fields are unexported, so a record can be
// read from outside the package but never changed.

package hospital

import "time"

// EpisodeRecord describes one sleep attack of a patient.
type EpisodeRecord struct {
	id         string
	patient    *Patient
	location   string
	occurredAt time.Time
	attendedBy Attender
	room       *Room // nil if the patient stayed in the hallway
}

// newEpisodeRecord is called by Attend with what the staff member knows.
// The Hospital adds the id, the time and the room when it registers it.
func newEpisodeRecord(p *Patient, location string, a Attender) EpisodeRecord {
	return EpisodeRecord{patient: p, location: location, attendedBy: a}
}

// ID returns the episode id, for example "E-001".
func (e EpisodeRecord) ID() string { return e.id }

// Patient returns the patient who fell asleep.
func (e EpisodeRecord) Patient() *Patient { return e.patient }

// Location returns where the attack happened.
func (e EpisodeRecord) Location() string { return e.location }

// OccurredAt returns when the attack was registered.
func (e EpisodeRecord) OccurredAt() time.Time { return e.occurredAt }

// AttendedBy returns the staff member who attended the patient.
func (e EpisodeRecord) AttendedBy() Attender { return e.attendedBy }

// Room returns the assigned room, or nil if there was none free.
func (e EpisodeRecord) Room() *Room { return e.room }
