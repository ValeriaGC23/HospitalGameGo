// This file defines the Attender interface. The Hospital keeps its staff as
// []Attender and sends patients to them without knowing if they are a
// Doctor or an Orderly.

package hospital

// Attender is a staff member who can attend a patient during a sleep attack.
type Attender interface {
	ID() string
	Name() string
	IsAvailable() bool
	// Attend returns the record of the attention, or an error if this staff
	// member cannot attend p.
	Attend(p *Patient, location string) (EpisodeRecord, error)

	// recordEpisode saves a finished record in the staff member's history.
	// It is unexported so that only the Hospital can write that history.
	recordEpisode(e EpisodeRecord)
}

// checkAttention holds the checks shared by every Attend method: the
// patient exists, the staff member is available and the patient is asleep
// in the hallway.
func checkAttention(a Attender, p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	if !a.IsAvailable() {
		return ErrAttenderUnavailable
	}
	if p.State() != AsleepInHallway {
		return ErrPatientNotInHallway
	}
	return nil
}
