// This file defines Doctor, a staff member that attends patients during
// sleep episodes and keeps the record of the episodes it attended.

package hospital

import "fmt"

// Doctor is a physician who can attend any patient, whatever the level of
// their narcolepsy.
type Doctor struct {
	Person
	available bool
	episodes  []EpisodeRecord
}

// Doctor satisfies Attender; the assignment fails to compile otherwise.
var _ Attender = (*Doctor)(nil)

// NewDoctor builds an available doctor with the given identity.
//
// It returns an error wrapping ErrInvalidID, ErrInvalidName or ErrInvalidAge
// when the identity is invalid.
func NewDoctor(id, name string, age int) (*Doctor, error) {
	person, err := newPerson(id, name, age)
	if err != nil {
		return nil, fmt.Errorf("new doctor %q: %w", id, err)
	}
	return &Doctor{Person: person, available: true}, nil
}

// IsAvailable reports whether the doctor can be dispatched.
func (d *Doctor) IsAvailable() bool { return d.available }

// SetAvailable marks the doctor as available or unavailable for dispatch.
func (d *Doctor) SetAvailable(available bool) { d.available = available }

// Attend attends p, who suffered a sleep attack at location.
//
// It returns ErrNilPatient when p is nil, ErrAttenderUnavailable when the
// doctor is unavailable and ErrPatientNotInHallway when p is not asleep in
// the hallway.
func (d *Doctor) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if err := checkAttention(d, p); err != nil {
		return EpisodeRecord{}, fmt.Errorf("doctor %q: %w", d.ID(), err)
	}
	return newEpisodeRecord(p, location, d), nil
}

// MyEpisodes returns a copy of the episodes this doctor attended, in
// registration order. The doctor keeps its own history, so the query does
// not need the Hospital.
func (d *Doctor) MyEpisodes() []EpisodeRecord {
	return append([]EpisodeRecord(nil), d.episodes...)
}

// recordEpisode implements Attender.
func (d *Doctor) recordEpisode(e EpisodeRecord) {
	d.episodes = append(d.episodes, e)
}
