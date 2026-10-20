// This file defines Person, the id, name and age shared by patients and
// staff. It is embedded by value, so these fields are written only once.

package hospital

import "strings"

// maxAge is the highest age accepted.
const maxAge = 130

// Person is the identity of a patient or staff member. It never changes
// after it is created, so its methods use value receivers.
type Person struct {
	id   string
	name string
	age  int
}

// newPerson checks the data and builds a Person. It returns ErrInvalidID,
// ErrInvalidName or ErrInvalidAge when a value is wrong.
func newPerson(id, name string, age int) (Person, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	if id == "" {
		return Person{}, ErrInvalidID
	}
	if name == "" {
		return Person{}, ErrInvalidName
	}
	if age < 0 || age > maxAge {
		return Person{}, ErrInvalidAge
	}
	return Person{id: id, name: name, age: age}, nil
}

// ID returns the person's identifier.
func (p Person) ID() string { return p.id }

// Name returns the person's name.
func (p Person) Name() string { return p.name }

// Age returns the person's age.
func (p Person) Age() int { return p.age }
