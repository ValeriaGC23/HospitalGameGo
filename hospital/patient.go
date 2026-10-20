// This file defines Patient and the ways its state can change:
//
//	Awake -> AsleepInHallway   SufferSleepAttack
//	AsleepInHallway -> AsleepInBed   moveToBed (only the Hospital)
//	asleep -> Awake   wakeUp (only the Hospital, through WakePatient)
//	Awake -> Discharged   Discharge

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

// NewPatient builds an awake patient. It returns an error if the id, name,
// age or level is not valid.
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

// Level returns the patient's narcolepsy level.
func (p *Patient) Level() NarcolepsyLevel { return p.level }

// State returns the patient's current state.
func (p *Patient) State() PatientState { return p.state }

// Location returns where the current attack happened.
func (p *Patient) Location() string { return p.location }

// SufferSleepAttack makes an awake patient fall asleep in the hallway. It
// returns ErrInvalidLocation, ErrPatientAsleep or ErrPatientDischarged, and
// in that case the patient does not change.
func (p *Patient) SufferSleepAttack(location string) error {
	location = strings.TrimSpace(location)
	if location == "" {
		return ErrInvalidLocation
	}
	if err := p.checkAwake(); err != nil {
		return err
	}
	p.state = AsleepInHallway
	p.location = location
	return nil
}

// Discharge sends an awake patient home. It returns ErrPatientAsleep or
// ErrPatientDischarged when that is not possible.
func (p *Patient) Discharge() error {
	if err := p.checkAwake(); err != nil {
		return err
	}
	p.state = Discharged
	return nil
}

// moveToBed marks the patient as asleep in a bed. It is unexported so that
// only the Hospital can do it, right after giving the patient a room. The
// Hospital checks first that the patient is asleep in the hallway.
func (p *Patient) moveToBed() {
	p.state = AsleepInBed
}

// wakeUp wakes a sleeping patient. It is unexported because the Hospital
// must also free the patient's bed (see Hospital.WakePatient).
func (p *Patient) wakeUp() error {
	if p.state == Awake {
		return ErrPatientAwake
	}
	if p.state == Discharged {
		return ErrPatientDischarged
	}
	p.state = Awake
	p.location = ""
	return nil
}

// checkAwake returns nil if the patient is awake, or the reason they are not.
func (p *Patient) checkAwake() error {
	if p.state == Discharged {
		return ErrPatientDischarged
	}
	if p.state != Awake {
		return ErrPatientAsleep
	}
	return nil
}
