// This file defines the enumerated state types of the domain: the state of a
// patient, the severity of their narcolepsy and the state of a room. Each type
// is backed by int with iota constants so that states are never loose strings.
// The zero value of every type is its natural initial state.

package hospital

import "fmt"

// PatientState is the current situation of a patient inside the hospital.
type PatientState int

// Patient states. Awake is the zero value, so a newly admitted patient starts
// awake without explicit initialization. Discharged is terminal: a discharged
// patient no longer takes part in sleep attacks or room assignment.
const (
	Awake PatientState = iota
	AsleepInHallway
	AsleepInBed
	Discharged
)

// String returns the name of the state, or "PatientState(n)" for a value
// outside the defined constants.
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
	default:
		return fmt.Sprintf("PatientState(%d)", int(s))
	}
}

// IsValid reports whether s is one of the defined PatientState constants.
func (s PatientState) IsValid() bool {
	return s >= Awake && s <= Discharged
}

// NarcolepsyLevel is the clinical severity of a patient's narcolepsy.
type NarcolepsyLevel int

// Narcolepsy levels, ordered from least to most severe so that levels can be
// compared with < and >. Severe is the level counted by SevereReport.
const (
	Mild NarcolepsyLevel = iota
	Moderate
	Severe
)

// String returns the name of the level, or "NarcolepsyLevel(n)" for a value
// outside the defined constants.
func (l NarcolepsyLevel) String() string {
	switch l {
	case Mild:
		return "Mild"
	case Moderate:
		return "Moderate"
	case Severe:
		return "Severe"
	default:
		return fmt.Sprintf("NarcolepsyLevel(%d)", int(l))
	}
}

// IsValid reports whether l is one of the defined NarcolepsyLevel constants.
func (l NarcolepsyLevel) IsValid() bool {
	return l >= Mild && l <= Severe
}

// RoomState is the occupancy state of a room.
type RoomState int

// Room states. Available is the zero value, so a newly registered room starts
// free without explicit initialization.
const (
	Available RoomState = iota
	Occupied
)

// String returns the name of the state, or "RoomState(n)" for a value outside
// the defined constants.
func (s RoomState) String() string {
	switch s {
	case Available:
		return "Available"
	case Occupied:
		return "Occupied"
	default:
		return fmt.Sprintf("RoomState(%d)", int(s))
	}
}

// IsValid reports whether s is one of the defined RoomState constants.
func (s RoomState) IsValid() bool {
	return s >= Available && s <= Occupied
}
