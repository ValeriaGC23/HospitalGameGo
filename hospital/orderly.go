// This file defines Orderly, a staff member who can carry sleeping patients
// to a room but cannot take charge of severe cases, which require a doctor.
// That restriction is what makes Orderly behave differently from Doctor
// behind the same Attender interface.

package hospital

import "fmt"

// Orderly is a hospital orderly who attends patients with Mild or Moderate
// narcolepsy.
type Orderly struct {
	Person
	available bool
	episodes  []EpisodeRecord
}

// Orderly satisfies Attender; the assignment fails to compile otherwise.
var _ Attender = (*Orderly)(nil)

// NewOrderly builds an available orderly with the given identity.
//
// It returns an error wrapping ErrInvalidID, ErrInvalidName or ErrInvalidAge
// when the identity is invalid.
func NewOrderly(id, name string, age int) (*Orderly, error) {
	person, err := newPerson(id, name, age)
	if err != nil {
		return nil, fmt.Errorf("new orderly %q: %w", id, err)
	}
	return &Orderly{Person: person, available: true}, nil
}

// IsAvailable reports whether the orderly can be dispatched.
func (o *Orderly) IsAvailable() bool { return o.available }

// SetAvailable marks the orderly as available or unavailable for dispatch.
func (o *Orderly) SetAvailable(available bool) { o.available = available }

// Attend attends p, who suffered a sleep attack at location.
//
// It returns ErrNilPatient when p is nil, ErrAttenderUnavailable when the
// orderly is unavailable, ErrPatientNotInHallway when p is not asleep in the
// hallway and ErrSevereNeedsDoctor when p has Severe narcolepsy.
func (o *Orderly) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if err := checkAttention(o, p); err != nil {
		return EpisodeRecord{}, fmt.Errorf("orderly %q: %w", o.ID(), err)
	}
	if p.Level() == Severe {
		return EpisodeRecord{}, fmt.Errorf("orderly %q: %w", o.ID(), ErrSevereNeedsDoctor)
	}
	return newEpisodeRecord(p, location, o), nil
}

// MyEpisodes returns a copy of the episodes this orderly attended, in
// registration order.
func (o *Orderly) MyEpisodes() []EpisodeRecord {
	return append([]EpisodeRecord(nil), o.episodes...)
}

// recordEpisode implements Attender.
func (o *Orderly) recordEpisode(e EpisodeRecord) {
	o.episodes = append(o.episodes, e)
}
