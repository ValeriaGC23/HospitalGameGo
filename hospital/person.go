// This file defines Person, the identity data shared by patients and staff.
// Person is embedded by value in Patient and Doctor so that id, name and age
// are declared once and their accessors are promoted to the outer types.

package hospital

import "strings"

// maxAge is the upper bound accepted for a person's age.
const maxAge = 130

// Person holds the identity of an individual. It is immutable after
// construction: there are no setters, so value receivers are safe and the
// methods are promoted to both T and *T of every type that embeds it.
type Person struct {
	id   string
	name string
	age  int
}

// newPerson validates and builds a Person. It is unexported because a Person
// only exists as part of a Patient or a Doctor.
//
// It returns ErrInvalidID or ErrInvalidName when id or name is blank, and
// ErrInvalidAge when age is outside [0, maxAge].
func newPerson(id, name string, age int) (Person, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	switch {
	case id == "":
		return Person{}, ErrInvalidID
	case name == "":
		return Person{}, ErrInvalidName
	case age < 0 || age > maxAge:
		return Person{}, ErrInvalidAge
	}
	return Person{id: id, name: name, age: age}, nil
}

// ID returns the person's identifier, unique within a hospital.
func (p Person) ID() string { return p.id }

// Name returns the person's name.
func (p Person) Name() string { return p.name }

// Age returns the person's age in years.
func (p Person) Age() int { return p.age }
