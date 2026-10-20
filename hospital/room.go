// This file defines Room. A room's state is calculated from its occupants,
// so it can never say "Available" while every bed is taken.

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

// NewRoom builds an empty room. It returns an error if the id is empty or
// the capacity is less than 1.
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

// ID returns the room identifier.
func (r *Room) ID() string { return r.id }

// Capacity returns the number of beds.
func (r *Room) Capacity() int { return r.capacity }

// State returns Occupied when every bed is taken, Available otherwise.
func (r *Room) State() RoomState {
	if len(r.occupants) >= r.capacity {
		return Occupied
	}
	return Available
}

// Occupants returns a copy of the patients in the room, so callers cannot
// change the room's own list.
func (r *Room) Occupants() []*Patient {
	return append([]*Patient(nil), r.occupants...)
}

// occupy puts p in a free bed. Only the Hospital calls it, after checking
// that the room is Available.
func (r *Room) occupy(p *Patient) {
	r.occupants = append(r.occupants, p)
}

// release frees p's bed. It returns ErrPatientNotInRoom if p is not here.
func (r *Room) release(p *Patient) error {
	for i, o := range r.occupants {
		if o == p {
			r.occupants = append(r.occupants[:i], r.occupants[i+1:]...)
			return nil
		}
	}
	return ErrPatientNotInRoom
}

// contains reports whether p is in the room.
func (r *Room) contains(p *Patient) bool {
	for _, o := range r.occupants {
		if o == p {
			return true
		}
	}
	return false
}
