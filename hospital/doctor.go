// This file defines Doctor, a staff member who can attend any patient.

package hospital

import "fmt"

// Doctor is a physician. Unlike an Orderly, a Doctor attends patients of
// every narcolepsy level.
type Doctor struct {
	Person
	available bool
	episodes  []EpisodeRecord
}

// This line does not compile if *Doctor stops implementing Attender.
var _ Attender = (*Doctor)(nil)

// NewDoctor builds an available doctor. It returns an error if the id, name
// or age is not valid.
func NewDoctor(id, name string, age int) (*Doctor, error) {
	person, err := newPerson(id, name, age)
	if err != nil {
		return nil, fmt.Errorf("new doctor %q: %w", id, err)
	}
	return &Doctor{Person: person, available: true}, nil
}

// IsAvailable reports whether the doctor can attend a patient now.
func (d *Doctor) IsAvailable() bool { return d.available }

// SetAvailable puts the doctor on or off duty.
func (d *Doctor) SetAvailable(available bool) { d.available = available }

// Attend attends p at location.
func (d *Doctor) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if err := checkAttention(d, p); err != nil {
		return EpisodeRecord{}, fmt.Errorf("doctor %q: %w", d.ID(), err)
	}
	return newEpisodeRecord(p, location, d), nil
}

// MyEpisodes returns a copy of the episodes this doctor attended. The doctor
// keeps its own history, so it does not need the Hospital.
func (d *Doctor) MyEpisodes() []EpisodeRecord {
	return append([]EpisodeRecord(nil), d.episodes...)
}

func (d *Doctor) recordEpisode(e EpisodeRecord) {
	d.episodes = append(d.episodes, e)
}
