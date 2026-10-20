# Hospital de los Costeños con Narcolepsia

Hospital management system written in Go that records patients with narcolepsy, their sudden sleep episodes, room assignment and the staff that attends them. Individual project for the Object-Oriented Programming course, Universidad EIA.

## Author

- Valeria (ValeriaGC23)

## Requirements

- Go 1.21 or later. Only the standard library is used.

## Run

```sh
go run .
```

The program runs the demonstration scenario:

1. Creates **My Hospital**, hires 2 doctors and 1 orderly, registers 3 rooms of capacity 1 and admits 5 patients (2 of them `Severe`).
2. Four patients suffer a sleep attack in different places. Three get a room; the fourth finds every room full, stays `AsleepInHallway` and the error is printed.
3. One patient wakes up and frees their bed; the patient left in the hallway is assigned to it through the normal `AssignRoom` flow.
4. Prints the four queries under `== ... ==` headings.

## Test

```sh
go test ./...
go test -v -cover ./hospital   # per-test output and coverage
```

13 tests, 85.5 % statement coverage of package `hospital`.

## Project structure

```
main.go                 scenario and printing only
hospital/
  types.go              PatientState, NarcolepsyLevel, RoomState (iota + String)
  errors.go             sentinel errors
  person.go             Person, embedded in Patient, Doctor and Orderly
  patient.go            Patient and its state transitions
  room.go               Room and its occupancy
  episode.go            EpisodeRecord
  attender.go           Attender interface
  doctor.go, orderly.go the two Attender implementations
  hospital.go           Hospital: admission, dispatch, room assignment, queries
  hospital_test.go      tests
```

## Design decisions

**Encapsulation.** Every struct has unexported fields and accessor methods. Package `hospital` holds all business logic and never prints; `main` builds the scenario and formats output.

**Composition.** `Person` (id, name, age) is embedded **by value** in `Patient`, `Doctor` and `Orderly`, so the fields are declared once and `ID()`, `Name()`, `Age()` are promoted. Value embedding means there is no nil `*Person` to check and each entity owns its own copy. `Person` is immutable, so its methods use value receivers and are promoted to both `T` and `*T`.

**Polymorphism.** The hospital stores its staff as `[]Attender` (`ID`, `Name`, `IsAvailable`, `Attend`) and dispatches without knowing the concrete type. The two implementations really behave differently: a `Doctor` attends any patient, while an `Orderly` refuses `Severe` patients with `ErrSevereNeedsDoctor`, and the hospital moves on to the next staff member. Dispatch is round-robin so that episodes are spread across the staff. `var _ Attender = (*Doctor)(nil)` checks each implementation at compile time.

**States as types.** `PatientState`, `NarcolepsyLevel` and `RoomState` are `int` types with `iota` constants and a `String()` method; no state is a loose string. The zero value of each is its natural initial state (`Awake`, `Mild`, `Available`).

**Invariants guarded by unexported methods.** Operations that must change two objects together are only reachable through the `Hospital`: `Room.occupy`/`release`, `Patient.moveToBed` and `Patient.wakeUp` are unexported, and the public entry points are `Hospital.AssignRoom` and `Hospital.WakePatient`. A room can therefore never be occupied by an awake patient. A room's state is derived from its occupants rather than stored, so the two cannot disagree.

**Errors.** Every method that can fail returns `error` last. Predictable failures (no room, duplicate id, invalid transition) are sentinel errors in `errors.go`, wrapped with context via `fmt.Errorf("...: %w", err)` and compared with `errors.Is`. There is no `panic`. When no room is free, `HandleSleepAttack` still registers the episode (the attack happened and was attended) and also returns the `ErrNoRoomAvailable` error.

**Queries return data.**

| Query | Returns |
|---|---|
| `(h *Hospital) PatientsInHallway() []*Patient` | patients asleep in the hallway, in admission order |
| `(d *Doctor) MyEpisodes() []EpisodeRecord` | the doctor's own history; the doctor does not need the hospital |
| `(h *Hospital) AssignRoom(p *Patient) (*Room, error)` | first available room, or `ErrNoRoomAvailable` without changing anything |
| `(h *Hospital) SevereReport() map[*Patient]int` | every `Severe` patient with their number of episodes today (0 included) |

Collections returned by methods are always copies, so callers cannot alter internal state.

**Slices over maps.** The hospital keeps patients, staff, rooms and episodes in slices so every query has a deterministic order; linear search is enough at this scale. `SevereReport` returns a map as required, and `main` sorts its keys before printing.

**Testable time.** The hospital has an injectable clock (`now func() time.Time`, `time.Now` by default). Episode timestamps come from it, and tests replace it with a fixed date so "episodes today" is deterministic. This is the only reason the tests use the internal package `hospital`.
