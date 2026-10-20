// This file lists the errors of the package. Callers check them with
// errors.Is; methods add context with fmt.Errorf("...: %w", err).

package hospital

import "errors"

// Errors for invalid data passed to constructors.
var (
	ErrInvalidID       = errors.New("id must not be empty")
	ErrInvalidName     = errors.New("name must not be empty")
	ErrInvalidAge      = errors.New("age out of range")
	ErrInvalidLevel    = errors.New("invalid narcolepsy level")
	ErrInvalidLocation = errors.New("location must not be empty")
	ErrInvalidCapacity = errors.New("capacity must be at least 1")
)

// Errors for state changes that are not allowed.
var (
	ErrPatientAsleep       = errors.New("patient is asleep")
	ErrPatientAwake        = errors.New("patient is awake")
	ErrPatientDischarged   = errors.New("patient is discharged")
	ErrPatientNotInHallway = errors.New("patient is not asleep in the hallway")
	ErrPatientNotInRoom    = errors.New("patient is not in the room")
)

// Errors returned by staff and by the hospital.
var (
	ErrNilPatient          = errors.New("patient must not be nil")
	ErrNilAttender         = errors.New("staff member must not be nil")
	ErrNilRoom             = errors.New("room must not be nil")
	ErrDuplicateID         = errors.New("duplicate id")
	ErrPatientNotAdmitted  = errors.New("patient is not admitted")
	ErrAttenderUnavailable = errors.New("staff member is not available")
	ErrSevereNeedsDoctor   = errors.New("severe patients must be attended by a doctor")
	ErrNoAttenderAvailable = errors.New("no staff member available")
	ErrNoRoomAvailable     = errors.New("no room available")
)
