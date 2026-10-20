# Hospital de los Costeños con Narcolepsia

Hospital management system written in Go that records patients with narcolepsy, their sudden sleep episodes, room assignment and the staff that attends them. Individual project for the Object-Oriented Programming course, Universidad EIA.

The project has three ways to run it:

| Command | What it does |
|---|---|
| `go run .` | The required demonstration scenario, printed in the terminal |
| `go run . -simulate` | Concurrent simulation in the terminal (bonus) |
| `go run . -web` | 2D game in the browser (bonus) |

## Author

- Valeria Gomez Cataño (ValeriaGC23)

## Requirements

- Go 1.21 or later. Only the standard library is used.
- Optional: a C compiler, only for the race detector (`-race`) on Windows. See [step 5](#5-check-for-data-races-optional).

## How to test the project

Open a terminal in the project folder (where `go.mod` is).

### 1. Run the tests

```sh
go test ./...
```

Expected: `ok  JuegoHospitalGo/hospital`. For the name of every test and the coverage:

```sh
go test -v -cover ./hospital
```

15 tests, about 90 % statement coverage of package `hospital`.

### 2. Run the demonstration scenario

```sh
go run .
```

It runs the scenario of the assignment, in this order:

1. Creates **My Hospital**, hires 2 doctors and 1 orderly, registers 3 rooms of capacity 1 and admits 5 patients (2 of them `Severe`).
2. Four patients suffer a sleep attack in different places. Three get a room; the fourth finds every room full, stays `AsleepInHallway` and the error is printed as a readable message.
3. One patient wakes up and frees their bed; the patient left in the hallway is assigned to it through the normal `AssignRoom` flow.
4. Prints the four queries under `== ... ==` headings.

### 3. Run the concurrent simulation (bonus)

```sh
go run . -simulate                 # 10 seconds
go run . -simulate -duration 30s
```

Every patient is a goroutine that falls asleep and wakes up at random times; the terminal prints each event with its time, and the four queries at the end.

### 4. Play the 2D game (bonus)

```sh
go run . -web
```

Open **http://localhost:8080** in a browser (use `-addr :9000` to change the port). Stop the server with `Ctrl+C`.

1. On the welcome screen, choose how many patients (3 to 10) and the mode, then press **Start**.
2. Patients walk around the cafeteria, waiting room, pharmacy and elevator, and **fall asleep by themselves**. The staff member who attended them walks over and carries them to a free bed, or to the hallway if every room is full.
3. Modes (the **Auto simulation** switch, which can be changed at any time):
   - **Auto on:** patients also **wake up by themselves**.
   - **Auto off (manual):** patients still fall asleep by themselves, but **you wake them up**: click a patient lying in a bed and press **Wake up**. They must have spent at least **5 seconds** in the room.
   - In both modes nobody stays asleep more than **30 seconds**.
4. Other actions: click an awake patient to trigger an attack; click a staff member to send them to rest (they walk to the staff room and only their head stays, click it to call them back). At least one doctor always stays on duty.
5. **Pause** freezes everything; **Restart** builds a new hospital and returns to the welcome screen.

Below the map: the rules of the game, the severe report and the cases solved by doctors. **📊 Data view** (`/dashboard.html`) shows the same data as a dashboard.

Things worth checking:
- The ring under each patient is their level: green Mild, amber Moderate, red Severe.
- An orderly never attends a Severe patient; a doctor does.
- When the 3 rooms are full, new patients stay in the hallway and move to a bed as soon as one is free.

### 5. Check for data races (optional)

The race detector needs cgo. On Linux and macOS it works directly; on Windows install a C compiler first (for example `winget install -e --id BrechtSanders.WinLibs.POSIX.UCRT`), open a new terminal and set `CGO_ENABLED`:

```sh
# Windows PowerShell: $env:CGO_ENABLED = "1"
go test -race ./...
go run -race . -simulate
```

Expected: no `WARNING: DATA RACE` in the output.

## Project structure

```
main.go                 scenario and printing only; chooses the mode with flags
simulation.go           goroutines of the simulation (-simulate and the game)
web.go                  web server and JSON API (-web)
web/index.html          2D game
web/dashboard.html      data dashboard
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
  snapshot.go           Snapshot: a copy of the state for concurrent readers
  hospital_test.go      tests
```

## Design decisions

**Encapsulation.** Every struct has unexported fields and accessor methods. Package `hospital` holds all business logic and never prints; `main` builds the scenario and formats output.

**Composition.** `Person` (id, name, age) is embedded **by value** in `Patient`, `Doctor` and `Orderly`, so the fields are declared once and `ID()`, `Name()`, `Age()` are promoted. Value embedding means there is no nil `*Person` to check and each entity owns its own copy. `Person` is immutable, so its methods use value receivers and are promoted to both `T` and `*T`.

**Polymorphism.** The hospital stores its staff as `[]Attender` and dispatches without knowing the concrete type. The two implementations really behave differently: a `Doctor` attends any patient, while an `Orderly` refuses `Severe` patients with `ErrSevereNeedsDoctor`, and the hospital moves on to the next staff member. Dispatch is round-robin so that episodes are spread across the staff. `var _ Attender = (*Doctor)(nil)` checks each implementation at compile time.

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

**Concurrency.** `Hospital` has a `sync.Mutex`; every exported method holds it for its whole duration, so two goroutines can never take the same bed. Exported methods never call each other (a `sync.Mutex` cannot be locked twice by the same goroutine); they share unexported helpers such as `assignRoom`. Goroutines that only need to read use `Hospital.Snapshot()`, a copy taken under the lock. They stop when a shared `stop` channel is closed, and `sync.WaitGroup` waits for all of them.

**Separation of logic and view.** The game (`web.go`, `web/`) only uses the public API of `hospital`. The game rules that are about presentation (5 seconds in a room before waking a patient by hand, the animations) live in the page; the hospital rules live in package `hospital`.
