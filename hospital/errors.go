// This file declares the sentinel errors of the package. Callers compare
// against them with errors.Is; methods that add context wrap them with %w.

package hospital

import "errors"

// Validation errors returned by constructors.
var (
	// ErrInvalidID is returned when an id is empty or only whitespace.
	ErrInvalidID = errors.New("id must not be empty")
	// ErrInvalidName is returned when a name is empty or only whitespace.
	ErrInvalidName = errors.New("name must not be empty")
	// ErrInvalidAge is returned when an age is outside [0, maxAge].
	ErrInvalidAge = errors.New("age out of range")
	// ErrInvalidLevel is returned when a NarcolepsyLevel is not a defined constant.
	ErrInvalidLevel = errors.New("invalid narcolepsy level")
	// ErrInvalidLocation is returned when a location is empty or only whitespace.
	ErrInvalidLocation = errors.New("location must not be empty")
	// ErrInvalidCapacity is returned when a room capacity is less than 1.
	ErrInvalidCapacity = errors.New("capacity must be at least 1")
	// ErrNilPatient is returned when a nil *Patient is passed where a
	// patient is required.
	ErrNilPatient = errors.New("patient must not be nil")
)

// Room occupancy errors returned by Room methods.
var (
	// ErrRoomFull is returned when a patient is placed in a room with no
	// free bed.
	ErrRoomFull = errors.New("room is full")
	// ErrPatientAlreadyInRoom is returned when a patient is placed in a
	// room they already occupy.
	ErrPatientAlreadyInRoom = errors.New("patient is already in the room")
	// ErrPatientNotInRoom is returned when a patient who is not an
	// occupant is released from a room.
	ErrPatientNotInRoom = errors.New("patient is not in the room")
)

// Attention errors returned by Attender implementations.
var (
	// ErrAttenderUnavailable is returned when an unavailable staff member is
	// asked to attend a patient.
	ErrAttenderUnavailable = errors.New("staff member is not available")
	// ErrSevereNeedsDoctor is returned when an Orderly is asked to attend a
	// patient with Severe narcolepsy.
	ErrSevereNeedsDoctor = errors.New("severe patients must be attended by a doctor")
)

// State transition errors returned by Patient methods.
var (
	// ErrPatientAsleep is returned when an operation requires an awake
	// patient and the patient is asleep.
	ErrPatientAsleep = errors.New("patient is asleep")
	// ErrPatientAwake is returned when an operation requires a sleeping
	// patient and the patient is awake.
	ErrPatientAwake = errors.New("patient is awake")
	// ErrPatientDischarged is returned for any state change on a
	// discharged patient.
	ErrPatientDischarged = errors.New("patient is discharged")
	// ErrPatientNotInHallway is returned when a patient is moved to a bed
	// without being asleep in the hallway.
	ErrPatientNotInHallway = errors.New("patient is not asleep in the hallway")
)
