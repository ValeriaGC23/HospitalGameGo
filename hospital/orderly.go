// This file defines Orderly, a staff member who carries patients to a room
// but cannot attend severe cases. That rule is what makes an Orderly behave
// differently from a Doctor behind the same Attender interface.

package hospital

import "fmt"

// Orderly attends patients with Mild or Moderate narcolepsy.
type Orderly struct {
	Person
	available bool
	episodes  []EpisodeRecord
}

// This line does not compile if *Orderly stops implementing Attender.
var _ Attender = (*Orderly)(nil)

// NewOrderly builds an available orderly. It returns an error if the id,
// name or age is not valid.
func NewOrderly(id, name string, age int) (*Orderly, error) {
	person, err := newPerson(id, name, age)
	if err != nil {
		return nil, fmt.Errorf("new orderly %q: %w", id, err)
	}
	return &Orderly{Person: person, available: true}, nil
}

// Role returns "Orderly".
func (o *Orderly) Role() string { return "Orderly" }

// IsAvailable reports whether the orderly can attend a patient now.
func (o *Orderly) IsAvailable() bool { return o.available }

// SetAvailable puts the orderly on or off duty.
func (o *Orderly) SetAvailable(available bool) { o.available = available }

// Attend attends p at location. It returns ErrSevereNeedsDoctor if p has
// Severe narcolepsy.
func (o *Orderly) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if err := checkAttention(o, p); err != nil {
		return EpisodeRecord{}, fmt.Errorf("orderly %q: %w", o.ID(), err)
	}
	if p.Level() == Severe {
		return EpisodeRecord{}, fmt.Errorf("orderly %q: %w", o.ID(), ErrSevereNeedsDoctor)
	}
	return newEpisodeRecord(p, location, o), nil
}

// MyEpisodes returns a copy of the episodes this orderly attended.
func (o *Orderly) MyEpisodes() []EpisodeRecord {
	return append([]EpisodeRecord(nil), o.episodes...)
}

func (o *Orderly) recordEpisode(e EpisodeRecord) {
	o.episodes = append(o.episodes, e)
}
