// This file defines Patient, a person admitted to the hospital whose state
// changes as sleep attacks occur and rooms are assigned.
//
// Valid transitions:
//
//	Awake --SufferSleepAttack--> AsleepInHallway --moveToBed--> AsleepInBed
//	AsleepInHallway, AsleepInBed --wakeUp--> Awake (via Hospital.WakePatient)
//	Awake --Discharge--> Discharged (terminal)
//
// Patient does not reference its Room: room occupancy is owned by the Room
// and coordinated by the Hospital, which avoids keeping two pointers in sync.

package hospital

import (
	"fmt"
	"strings"
)

// Patient is a person with narcolepsy admitted to the hospital.
type Patient struct {
	Person
	level    NarcolepsyLevel
	state    PatientState
	location string // where the current attack happened; empty while awake
}

// NewPatient builds an awake patient with the given identity and level.
//
// It returns an error wrapping ErrInvalidID, ErrInvalidName or ErrInvalidAge
// when the identity is invalid, or ErrInvalidLevel when level is not a
// defined NarcolepsyLevel.
func NewPatient(id, name string, age int, level NarcolepsyLevel) (*Patient, error) {
	person, err := newPerson(id, name, age)
	if err != nil {
		return nil, fmt.Errorf("new patient %q: %w", id, err)
	}
	if !level.IsValid() {
		return nil, fmt.Errorf("new patient %q: %w", id, ErrInvalidLevel)
	}
	return &Patient{Person: person, level: level}, nil
}

// Level returns the severity of the patient's narcolepsy.
func (p *Patient) Level() NarcolepsyLevel { return p.level }

// State returns the patient's current state.
func (p *Patient) State() PatientState { return p.state }

// Location returns where the patient's current sleep attack happened, or an
// empty string while the patient is awake or discharged.
func (p *Patient) Location() string { return p.location }

// SufferSleepAttack puts an awake patient to sleep in the hallway at location.
//
// It returns ErrInvalidLocation when location is blank, ErrPatientAsleep when
// the patient is already asleep and ErrPatientDischarged when the patient has
// been discharged. On error the patient is left unchanged.
func (p *Patient) SufferSleepAttack(location string) error {
	location = strings.TrimSpace(location)
	if location == "" {
		return ErrInvalidLocation
	}
	if err := p.requireAwake(); err != nil {
		return err
	}
	p.state = AsleepInHallway
	p.location = location
	return nil
}

// wakeUp returns a sleeping patient to the awake state and clears the
// location. It is unexported because waking a patient in a bed must also
// release the room; Hospital.WakePatient performs both steps together.
//
// It returns ErrPatientAwake when the patient is already awake and
// ErrPatientDischarged when the patient has been discharged.
func (p *Patient) wakeUp() error {
	switch p.state {
	case Awake:
		return ErrPatientAwake
	case Discharged:
		return ErrPatientDischarged
	}
	p.state = Awake
	p.location = ""
	return nil
}

// Discharge moves an awake patient to the terminal Discharged state.
//
// It returns ErrPatientAsleep when the patient is asleep and
// ErrPatientDischarged when the patient was already discharged.
func (p *Patient) Discharge() error {
	if err := p.requireAwake(); err != nil {
		return err
	}
	p.state = Discharged
	return nil
}

// moveToBed marks a patient asleep in the hallway as asleep in a bed. It is
// unexported so that only the Hospital, after occupying a Room, can perform
// this transition.
//
// It returns ErrPatientNotInHallway when the patient is in any other state.
func (p *Patient) moveToBed() error {
	if p.state != AsleepInHallway {
		return ErrPatientNotInHallway
	}
	p.state = AsleepInBed
	return nil
}

// requireAwake returns nil when the patient is awake, ErrPatientDischarged
// when discharged and ErrPatientAsleep otherwise.
func (p *Patient) requireAwake() error {
	switch p.state {
	case Awake:
		return nil
	case Discharged:
		return ErrPatientDischarged
	default:
		return ErrPatientAsleep
	}
}
