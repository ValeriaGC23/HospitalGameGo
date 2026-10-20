// This file defines Room, a hospital room with a fixed capacity that can be
// occupied by and released from patients.
//
// The room state is derived from its occupants instead of being stored, so
// it can never disagree with them: a room is Occupied when it is full and
// Available while it has at least one free bed.

package hospital

import (
	"fmt"
	"strings"
)

// Room is a hospital room with a fixed number of beds.
type Room struct {
	id        string
	capacity  int
	occupants []*Patient
}

// NewRoom builds an empty room with the given id and number of beds.
//
// It returns an error wrapping ErrInvalidID when id is blank, or
// ErrInvalidCapacity when capacity is less than 1.
func NewRoom(id string, capacity int) (*Room, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("new room: %w", ErrInvalidID)
	}
	if capacity < 1 {
		return nil, fmt.Errorf("new room %q: %w", id, ErrInvalidCapacity)
	}
	return &Room{id: id, capacity: capacity}, nil
}

// ID returns the room identifier, unique within a hospital.
func (r *Room) ID() string { return r.id }

// Capacity returns the number of beds in the room.
func (r *Room) Capacity() int { return r.capacity }

// State returns Occupied when every bed is taken and Available otherwise.
func (r *Room) State() RoomState {
	if len(r.occupants) >= r.capacity {
		return Occupied
	}
	return Available
}

// Occupants returns a copy of the patients currently in the room, so callers
// cannot modify the room's internal slice.
func (r *Room) Occupants() []*Patient {
	return append([]*Patient(nil), r.occupants...)
}

// occupy places p in a free bed. It is unexported so that only the Hospital
// can occupy a room, together with moving the patient to the AsleepInBed
// state.
//
// It returns ErrNilPatient when p is nil, ErrPatientAlreadyInRoom when p is
// already an occupant and ErrRoomFull when no bed is free. On error the room
// is left unchanged.
func (r *Room) occupy(p *Patient) error {
	switch {
	case p == nil:
		return ErrNilPatient
	case r.contains(p):
		return ErrPatientAlreadyInRoom
	case r.State() == Occupied:
		return ErrRoomFull
	}
	r.occupants = append(r.occupants, p)
	return nil
}

// release frees the bed held by p. It is unexported for the same reason as
// occupy.
//
// It returns ErrPatientNotInRoom when p is not an occupant.
func (r *Room) release(p *Patient) error {
	for i, o := range r.occupants {
		if o == p {
			r.occupants = append(r.occupants[:i], r.occupants[i+1:]...)
			return nil
		}
	}
	return ErrPatientNotInRoom
}

// contains reports whether p is an occupant of the room. Patients are
// compared by pointer identity.
func (r *Room) contains(p *Patient) bool {
	for _, o := range r.occupants {
		if o == p {
			return true
		}
	}
	return false
}
