// This file defines the state types: patient state, narcolepsy level and
// room state. Each one is an int with iota constants, never a loose string.

package hospital

import "fmt"

// PatientState is the current situation of a patient.
type PatientState int

// Patient states. Awake is the zero value, so a new patient starts awake.
const (
	Awake PatientState = iota
	AsleepInHallway
	AsleepInBed
	Discharged
)

// String returns the name of the state.
func (s PatientState) String() string {
	switch s {
	case Awake:
		return "Awake"
	case AsleepInHallway:
		return "AsleepInHallway"
	case AsleepInBed:
		return "AsleepInBed"
	case Discharged:
		return "Discharged"
	}
	return fmt.Sprintf("PatientState(%d)", int(s))
}

// NarcolepsyLevel is how severe a patient's narcolepsy is, from Mild to
// Severe, so levels can be compared with < and >.
type NarcolepsyLevel int

// Narcolepsy levels.
const (
	Mild NarcolepsyLevel = iota
	Moderate
	Severe
)

// String returns the name of the level.
func (l NarcolepsyLevel) String() string {
	switch l {
	case Mild:
		return "Mild"
	case Moderate:
		return "Moderate"
	case Severe:
		return "Severe"
	}
	return fmt.Sprintf("NarcolepsyLevel(%d)", int(l))
}

// IsValid reports whether l is one of the defined levels. NewPatient uses it
// to reject values such as NarcolepsyLevel(9).
func (l NarcolepsyLevel) IsValid() bool {
	return l >= Mild && l <= Severe
}

// RoomState tells whether a room still has a free bed.
type RoomState int

// Room states.
const (
	Available RoomState = iota
	Occupied
)

// String returns the name of the state.
func (s RoomState) String() string {
	switch s {
	case Available:
		return "Available"
	case Occupied:
		return "Occupied"
	}
	return fmt.Sprintf("RoomState(%d)", int(s))
}
