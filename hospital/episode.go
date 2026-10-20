// This file defines EpisodeRecord, the record of a single sleep attack:
// the patient, where and when it happened, who attended it and the room
// assigned afterwards.
//
// EpisodeRecord is a value type: it is copied into the hospital history and
// into the history of the staff member who attended it, and it is never
// modified after registration. Its fields are unexported so that callers
// outside the package can read a record but cannot alter it.

package hospital

import "time"

// EpisodeRecord describes one sleep attack of a patient.
type EpisodeRecord struct {
	id         string
	patient    *Patient
	location   string
	occurredAt time.Time
	attendedBy Attender
	room       *Room // nil when the patient stayed in the hallway
}

// newEpisodeRecord builds a partial record with the data known by the staff
// member at attention time. The Hospital sets id, occurredAt and room when it
// registers the episode, so that ids and timestamps come from a single source
// (its counter and its clock).
func newEpisodeRecord(p *Patient, location string, a Attender) EpisodeRecord {
	return EpisodeRecord{patient: p, location: location, attendedBy: a}
}

// ID returns the episode identifier assigned by the Hospital (e.g. "E-001").
func (e EpisodeRecord) ID() string { return e.id }

// Patient returns the patient who suffered the attack.
func (e EpisodeRecord) Patient() *Patient { return e.patient }

// Location returns where the attack happened.
func (e EpisodeRecord) Location() string { return e.location }

// OccurredAt returns when the episode was registered by the Hospital.
func (e EpisodeRecord) OccurredAt() time.Time { return e.occurredAt }

// AttendedBy returns the staff member who attended the patient.
func (e EpisodeRecord) AttendedBy() Attender { return e.attendedBy }

// Room returns the room assigned to the patient, or nil when no room was
// available and the patient stayed asleep in the hallway.
func (e EpisodeRecord) Room() *Room { return e.room }
