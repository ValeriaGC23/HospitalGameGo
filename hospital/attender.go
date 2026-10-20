// This file declares Attender, the behavior shared by every kind of staff
// member that can attend a patient during a sleep attack. The Hospital keeps
// its staff as []Attender and dispatches without knowing the concrete type.

package hospital

// Attender is a staff member that can attend a patient during a sleep attack.
//
// IsAvailable extends the minimum required interface so that the Hospital can
// choose whom to dispatch without type assertions. The unexported
// recordEpisode method restricts implementations to this package, which
// guarantees that only the Hospital writes to a staff member's history.
type Attender interface {
	// ID returns the staff member's identifier, unique within a hospital.
	ID() string
	// Name returns the staff member's name.
	Name() string
	// IsAvailable reports whether the staff member can attend a patient now.
	IsAvailable() bool
	// Attend attends p, who suffered a sleep attack at location, and returns
	// the resulting record. The record has no id, time or room yet: the
	// Hospital completes them when it registers the episode.
	Attend(p *Patient, location string) (EpisodeRecord, error)

	// recordEpisode appends a registered episode to the staff member's
	// history. Only the Hospital calls it, once the record is complete.
	recordEpisode(e EpisodeRecord)
}

// checkAttention validates the preconditions shared by every Attend
// implementation.
//
// It returns ErrNilPatient when p is nil, ErrAttenderUnavailable when a is
// not available and ErrPatientNotInHallway when p is not asleep in the
// hallway, which is the only state in which a patient needs attention.
func checkAttention(a Attender, p *Patient) error {
	switch {
	case p == nil:
		return ErrNilPatient
	case !a.IsAvailable():
		return ErrAttenderUnavailable
	case p.State() != AsleepInHallway:
		return ErrPatientNotInHallway
	}
	return nil
}
