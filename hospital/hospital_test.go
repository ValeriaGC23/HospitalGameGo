// This file contains the tests for package hospital. They exercise the
// public API; the only internal access is replacing the hospital clock
// (h.now), which is indispensable for deterministic "today" checks in
// SevereReport and is why the tests live in package hospital.

package hospital

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// day1 is the fixed "today" used by the test clock.
var day1 = time.Date(2026, time.October, 19, 10, 0, 0, 0, time.UTC)

// fixture is a hospital populated through the public API.
type fixture struct {
	h       *Hospital
	doctors []*Doctor
	orderly *Orderly
	rooms   []*Room
}

// newFixture builds a hospital with a fixed clock, the given number of
// doctors (plus one orderly when withOrderly is true) and rooms of capacity 1.
func newFixture(t *testing.T, doctors, rooms int, withOrderly bool) *fixture {
	t.Helper()
	h, err := NewHospital("Test Hospital")
	if err != nil {
		t.Fatalf("NewHospital: %v", err)
	}
	h.now = func() time.Time { return day1 }
	f := &fixture{h: h}
	for i := 0; i < doctors; i++ {
		d, err := NewDoctor(id("D", i), "Doctor", 45)
		if err != nil {
			t.Fatalf("NewDoctor: %v", err)
		}
		if err := h.HireDoctor(d); err != nil {
			t.Fatalf("HireDoctor: %v", err)
		}
		f.doctors = append(f.doctors, d)
	}
	if withOrderly {
		o, err := NewOrderly("O-1", "Orderly", 30)
		if err != nil {
			t.Fatalf("NewOrderly: %v", err)
		}
		if err := h.HireAttender(o); err != nil {
			t.Fatalf("HireAttender: %v", err)
		}
		f.orderly = o
	}
	for i := 0; i < rooms; i++ {
		r, err := NewRoom(id("R", i), 1)
		if err != nil {
			t.Fatalf("NewRoom: %v", err)
		}
		if err := h.RegisterRoom(r); err != nil {
			t.Fatalf("RegisterRoom: %v", err)
		}
		f.rooms = append(f.rooms, r)
	}
	return f
}

// admit creates and admits a patient with the given id and level.
func (f *fixture) admit(t *testing.T, pid string, level NarcolepsyLevel) *Patient {
	t.Helper()
	p, err := NewPatient(pid, "Patient "+pid, 30, level)
	if err != nil {
		t.Fatalf("NewPatient: %v", err)
	}
	if err := f.h.AdmitPatient(p); err != nil {
		t.Fatalf("AdmitPatient: %v", err)
	}
	return p
}

// asleepInHallway admits a patient and puts them to sleep without
// dispatching staff or assigning a room.
func (f *fixture) asleepInHallway(t *testing.T, pid string, level NarcolepsyLevel) *Patient {
	t.Helper()
	p := f.admit(t, pid, level)
	if err := p.SufferSleepAttack("Hallway " + pid); err != nil {
		t.Fatalf("SufferSleepAttack: %v", err)
	}
	return p
}

// id builds ids such as "D-1" from a prefix and a zero-based index.
func id(prefix string, i int) string {
	return prefix + "-" + string(rune('1'+i))
}

func TestAssignRoomWithFreeRoom(t *testing.T) {
	f := newFixture(t, 1, 1, false)
	p := f.asleepInHallway(t, "P-1", Mild)

	room, err := f.h.AssignRoom(p)

	if err != nil {
		t.Fatalf("AssignRoom: unexpected error %v", err)
	}
	if room != f.rooms[0] {
		t.Fatalf("AssignRoom: got room %v, want %s", room, f.rooms[0].ID())
	}
	if got := room.State(); got != Occupied {
		t.Errorf("room state = %v, want %v", got, Occupied)
	}
	if got := p.State(); got != AsleepInBed {
		t.Errorf("patient state = %v, want %v", got, AsleepInBed)
	}
	if occ := room.Occupants(); len(occ) != 1 || occ[0] != p {
		t.Errorf("room occupants = %v, want [%s]", occ, p.ID())
	}
}

func TestAssignRoomWithoutFreeRoom(t *testing.T) {
	f := newFixture(t, 1, 1, false)
	first := f.asleepInHallway(t, "P-1", Mild)
	if _, err := f.h.AssignRoom(first); err != nil {
		t.Fatalf("AssignRoom(first): %v", err)
	}
	second := f.asleepInHallway(t, "P-2", Mild)
	roomsBefore := len(f.h.Rooms())

	room, err := f.h.AssignRoom(second)

	if !errors.Is(err, ErrNoRoomAvailable) {
		t.Fatalf("AssignRoom error = %v, want %v", err, ErrNoRoomAvailable)
	}
	if room != nil {
		t.Errorf("AssignRoom room = %s, want nil", room.ID())
	}
	if got := second.State(); got != AsleepInHallway {
		t.Errorf("patient state = %v, want %v", got, AsleepInHallway)
	}
	if got := len(f.h.Rooms()); got != roomsBefore {
		t.Errorf("rooms = %d, want %d (no room must be created)", got, roomsBefore)
	}
}

func TestAssignRoomRejectsInvalidPatients(t *testing.T) {
	f := newFixture(t, 1, 1, false)
	awake := f.admit(t, "P-1", Mild)
	stranger, err := NewPatient("P-9", "Stranger", 30, Mild)
	if err != nil {
		t.Fatalf("NewPatient: %v", err)
	}

	tests := []struct {
		name    string
		patient *Patient
		wantErr error
	}{
		{"nil patient", nil, ErrNilPatient},
		{"not admitted", stranger, ErrPatientNotAdmitted},
		{"awake patient", awake, ErrPatientNotInHallway},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room, err := f.h.AssignRoom(tt.patient)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if room != nil {
				t.Errorf("room = %s, want nil", room.ID())
			}
			if got := f.rooms[0].State(); got != Available {
				t.Errorf("room state = %v, want %v", got, Available)
			}
		})
	}
}

func TestPatientsInHallway(t *testing.T) {
	f := newFixture(t, 2, 1, false)
	inBed := f.admit(t, "P-1", Mild)
	f.admit(t, "P-2", Mild) // stays awake: must not be listed
	inHallway := f.admit(t, "P-3", Severe)
	if _, err := f.h.HandleSleepAttack(inBed, "Lobby"); err != nil {
		t.Fatalf("HandleSleepAttack(inBed): %v", err)
	}
	if _, err := f.h.HandleSleepAttack(inHallway, "Cafeteria"); !errors.Is(err, ErrNoRoomAvailable) {
		t.Fatalf("HandleSleepAttack(inHallway) error = %v, want %v", err, ErrNoRoomAvailable)
	}

	got := f.h.PatientsInHallway()

	if len(got) != 1 || got[0] != inHallway {
		t.Fatalf("PatientsInHallway = %v, want [%s]", got, inHallway.ID())
	}
}

func TestMyEpisodes(t *testing.T) {
	f := newFixture(t, 2, 3, false)
	p1 := f.admit(t, "P-1", Mild)
	p2 := f.admit(t, "P-2", Moderate)
	p3 := f.admit(t, "P-3", Severe)
	for _, p := range []*Patient{p1, p2, p3} {
		if _, err := f.h.HandleSleepAttack(p, "Lobby"); err != nil {
			t.Fatalf("HandleSleepAttack(%s): %v", p.ID(), err)
		}
	}

	// Round-robin dispatch: D-1 attends P-1 and P-3, D-2 attends P-2.
	tests := []struct {
		doctor *Doctor
		want   []*Patient
	}{
		{f.doctors[0], []*Patient{p1, p3}},
		{f.doctors[1], []*Patient{p2}},
	}
	for _, tt := range tests {
		t.Run(tt.doctor.ID(), func(t *testing.T) {
			episodes := tt.doctor.MyEpisodes()
			if len(episodes) != len(tt.want) {
				t.Fatalf("episodes = %d, want %d", len(episodes), len(tt.want))
			}
			for i, e := range episodes {
				if e.Patient() != tt.want[i] {
					t.Errorf("episode %d patient = %s, want %s", i, e.Patient().ID(), tt.want[i].ID())
				}
				if e.AttendedBy() != Attender(tt.doctor) {
					t.Errorf("episode %d attended by %s, want %s", i, e.AttendedBy().ID(), tt.doctor.ID())
				}
				if e.ID() == "" || e.Room() == nil || !e.OccurredAt().Equal(day1) {
					t.Errorf("episode %d incomplete: id=%q room=%v at=%v", i, e.ID(), e.Room(), e.OccurredAt())
				}
			}
		})
	}

	episodes := f.doctors[0].MyEpisodes()
	episodes[0] = EpisodeRecord{}
	if f.doctors[0].MyEpisodes()[0].Patient() != p1 {
		t.Error("MyEpisodes must return a copy of the history")
	}
}

func TestSevereReport(t *testing.T) {
	f := newFixture(t, 1, 3, false)
	twice := f.admit(t, "P-1", Severe)
	never := f.admit(t, "P-2", Severe)
	mild := f.admit(t, "P-3", Mild)

	// Episode on the previous day: must not be counted.
	f.h.now = func() time.Time { return day1.AddDate(0, 0, -1) }
	attackAndWake(t, f.h, twice)
	f.h.now = func() time.Time { return day1 }
	attackAndWake(t, f.h, twice)
	attackAndWake(t, f.h, twice)
	attackAndWake(t, f.h, mild)

	report := f.h.SevereReport()

	want := map[*Patient]int{twice: 2, never: 0}
	if len(report) != len(want) {
		t.Fatalf("report has %d patients, want %d: %v", len(report), len(want), report)
	}
	for p, n := range want {
		if got, ok := report[p]; !ok || got != n {
			t.Errorf("report[%s] = %d (present %t), want %d", p.ID(), got, ok, n)
		}
	}
}

// attackAndWake runs a full sleep attack for p and wakes them up, freeing
// the room for the next attack.
func attackAndWake(t *testing.T, h *Hospital, p *Patient) {
	t.Helper()
	if _, err := h.HandleSleepAttack(p, "Lobby"); err != nil {
		t.Fatalf("HandleSleepAttack(%s): %v", p.ID(), err)
	}
	if err := h.WakePatient(p); err != nil {
		t.Fatalf("WakePatient(%s): %v", p.ID(), err)
	}
}

func TestHandleSleepAttackDispatch(t *testing.T) {
	f := newFixture(t, 1, 3, true)
	severe := f.admit(t, "P-1", Severe)
	mild := f.admit(t, "P-2", Mild)
	other := f.admit(t, "P-3", Moderate)

	tests := []struct {
		name    string
		patient *Patient
		want    Attender
	}{
		{"doctor takes first turn", severe, f.doctors[0]},
		{"orderly takes next turn", mild, f.orderly},
		{"rotation returns to doctor", other, f.doctors[0]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := f.h.HandleSleepAttack(tt.patient, "Lobby")
			if err != nil {
				t.Fatalf("HandleSleepAttack: %v", err)
			}
			if e.AttendedBy() != tt.want {
				t.Errorf("attended by %s, want %s", e.AttendedBy().ID(), tt.want.ID())
			}
		})
	}
}

// TestSevereWithoutDoctorIsNotAttended checks that an Orderly never attends
// a Severe patient, even when it is the only staff member available.
func TestSevereWithoutDoctorIsNotAttended(t *testing.T) {
	f := newFixture(t, 1, 1, true)
	f.doctors[0].SetAvailable(false)
	severe := f.admit(t, "P-1", Severe)

	_, err := f.h.HandleSleepAttack(severe, "Lobby")

	if !errors.Is(err, ErrNoAttenderAvailable) {
		t.Fatalf("error = %v, want %v", err, ErrNoAttenderAvailable)
	}
	if len(f.h.Episodes()) != 0 {
		t.Errorf("episodes = %d, want 0", len(f.h.Episodes()))
	}
	if got := severe.State(); got != AsleepInHallway {
		t.Errorf("patient state = %v, want %v", got, AsleepInHallway)
	}
}

func TestWakePatientReleasesRoom(t *testing.T) {
	f := newFixture(t, 1, 1, false)
	inBed := f.admit(t, "P-1", Mild)
	waiting := f.admit(t, "P-2", Mild)
	if _, err := f.h.HandleSleepAttack(inBed, "Lobby"); err != nil {
		t.Fatalf("HandleSleepAttack(inBed): %v", err)
	}
	if _, err := f.h.HandleSleepAttack(waiting, "Yard"); !errors.Is(err, ErrNoRoomAvailable) {
		t.Fatalf("HandleSleepAttack(waiting) error = %v, want %v", err, ErrNoRoomAvailable)
	}

	if err := f.h.WakePatient(inBed); err != nil {
		t.Fatalf("WakePatient: %v", err)
	}
	if got := inBed.State(); got != Awake {
		t.Errorf("woken patient state = %v, want %v", got, Awake)
	}
	if got := f.rooms[0].State(); got != Available {
		t.Fatalf("room state after wake = %v, want %v", got, Available)
	}
	room, err := f.h.AssignRoom(waiting)
	if err != nil || room != f.rooms[0] {
		t.Fatalf("AssignRoom(waiting) = %v, %v; want %s, nil", room, err, f.rooms[0].ID())
	}
	if err := f.h.WakePatient(inBed); !errors.Is(err, ErrPatientAwake) {
		t.Errorf("WakePatient(awake) error = %v, want %v", err, ErrPatientAwake)
	}
}

func TestPatientTransitions(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(p *Patient)
		action  func(p *Patient) error
		wantErr error
		want    PatientState
	}{
		{"attack while awake", nil, attack, nil, AsleepInHallway},
		{"attack while asleep", func(p *Patient) { _ = attack(p) }, attack, ErrPatientAsleep, AsleepInHallway},
		{"attack at blank location", nil, func(p *Patient) error { return p.SufferSleepAttack("  ") }, ErrInvalidLocation, Awake},
		{"discharge while awake", nil, (*Patient).Discharge, nil, Discharged},
		{"discharge while asleep", func(p *Patient) { _ = attack(p) }, (*Patient).Discharge, ErrPatientAsleep, AsleepInHallway},
		{"attack after discharge", func(p *Patient) { _ = p.Discharge() }, attack, ErrPatientDischarged, Discharged},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPatient("P-1", "Patient", 30, Mild)
			if err != nil {
				t.Fatalf("NewPatient: %v", err)
			}
			if tt.setup != nil {
				tt.setup(p)
			}
			if err := tt.action(p); !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if got := p.State(); got != tt.want {
				t.Errorf("state = %v, want %v", got, tt.want)
			}
		})
	}
}

// attack puts p to sleep at a fixed location.
func attack(p *Patient) error { return p.SufferSleepAttack("Lobby") }

func TestConstructorsValidate(t *testing.T) {
	tests := []struct {
		name    string
		build   func() error
		wantErr error
	}{
		{"patient blank id", func() error { _, err := NewPatient(" ", "Ana", 30, Mild); return err }, ErrInvalidID},
		{"patient blank name", func() error { _, err := NewPatient("P-1", "", 30, Mild); return err }, ErrInvalidName},
		{"patient negative age", func() error { _, err := NewPatient("P-1", "Ana", -1, Mild); return err }, ErrInvalidAge},
		{"patient invalid level", func() error { _, err := NewPatient("P-1", "Ana", 30, NarcolepsyLevel(9)); return err }, ErrInvalidLevel},
		{"doctor age too high", func() error { _, err := NewDoctor("D-1", "Doc", maxAge+1); return err }, ErrInvalidAge},
		{"orderly blank name", func() error { _, err := NewOrderly("O-1", " ", 30); return err }, ErrInvalidName},
		{"room zero capacity", func() error { _, err := NewRoom("R-1", 0); return err }, ErrInvalidCapacity},
		{"hospital blank name", func() error { _, err := NewHospital(""); return err }, ErrInvalidName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.build(); !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDuplicateIDsRejected(t *testing.T) {
	f := newFixture(t, 1, 1, false)
	f.admit(t, "P-1", Mild)
	dupPatient, _ := NewPatient("P-1", "Other", 30, Mild)
	dupDoctor, _ := NewDoctor(f.doctors[0].ID(), "Other", 40)
	dupRoom, _ := NewRoom(f.rooms[0].ID(), 1)

	if err := f.h.AdmitPatient(dupPatient); !errors.Is(err, ErrDuplicateID) {
		t.Errorf("AdmitPatient error = %v, want %v", err, ErrDuplicateID)
	}
	if err := f.h.HireDoctor(dupDoctor); !errors.Is(err, ErrDuplicateID) {
		t.Errorf("HireDoctor error = %v, want %v", err, ErrDuplicateID)
	}
	if err := f.h.RegisterRoom(dupRoom); !errors.Is(err, ErrDuplicateID) {
		t.Errorf("RegisterRoom error = %v, want %v", err, ErrDuplicateID)
	}
}

// TestConcurrentUse runs many goroutines against one hospital at the same
// time. Run it with "go test -race ./..." to also check for data races.
// Without the mutex, two goroutines could take the same free bed.
func TestConcurrentUse(t *testing.T) {
	f := newFixture(t, 2, 3, true)
	var patients []*Patient
	for i := 0; i < 8; i++ {
		patients = append(patients, f.admit(t, id("P", i), NarcolepsyLevel(i%3)))
	}

	var wg sync.WaitGroup
	for _, p := range patients {
		wg.Add(1)
		go func(p *Patient) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				_, _ = f.h.HandleSleepAttack(p, "Lobby")
				_ = f.h.Snapshot()
				_ = f.h.WakePatient(p)
			}
		}(p)
	}
	wg.Wait()

	s := f.h.Snapshot()
	for _, r := range s.Rooms {
		if len(r.Occupants) > r.Capacity {
			t.Errorf("room %s has %d occupants, capacity %d", r.ID, len(r.Occupants), r.Capacity)
		}
	}
	for _, p := range s.Patients {
		if p.State != Awake || p.RoomID != "" {
			t.Errorf("patient %s ended as %v in room %q, want Awake and no room", p.ID, p.State, p.RoomID)
		}
	}
}

func TestSnapshot(t *testing.T) {
	f := newFixture(t, 1, 1, false)
	inBed := f.admit(t, "P-1", Severe)
	inHallway := f.admit(t, "P-2", Mild)
	if _, err := f.h.HandleSleepAttack(inBed, "Lobby"); err != nil {
		t.Fatalf("HandleSleepAttack(inBed): %v", err)
	}
	_, _ = f.h.HandleSleepAttack(inHallway, "Yard")

	s := f.h.Snapshot()

	if len(s.Patients) != 2 || s.Patients[0].RoomID != f.rooms[0].ID() || s.Patients[1].State != AsleepInHallway {
		t.Errorf("patients = %+v", s.Patients)
	}
	if len(s.Episodes) != 2 || s.Episodes[1].RoomID != "" {
		t.Errorf("episodes = %+v", s.Episodes)
	}
	if s.Staff[0].Role != "Doctor" || s.Staff[0].Episodes != 2 {
		t.Errorf("staff = %+v", s.Staff)
	}
	if s.SevereToday[inBed.ID()] != 1 || len(s.SevereToday) != 1 {
		t.Errorf("severe = %v", s.SevereToday)
	}
	if err := f.h.SetAttenderAvailable("D-1", false); err != nil || f.h.Snapshot().Staff[0].Available {
		t.Errorf("SetAttenderAvailable: %v", err)
	}
}

func TestStateStrings(t *testing.T) {
	tests := []struct {
		got  string
		want string
	}{
		{Awake.String(), "Awake"},
		{AsleepInHallway.String(), "AsleepInHallway"},
		{AsleepInBed.String(), "AsleepInBed"},
		{Discharged.String(), "Discharged"},
		{PatientState(42).String(), "PatientState(42)"},
		{Mild.String(), "Mild"},
		{Moderate.String(), "Moderate"},
		{Severe.String(), "Severe"},
		{NarcolepsyLevel(-1).String(), "NarcolepsyLevel(-1)"},
		{Available.String(), "Available"},
		{Occupied.String(), "Occupied"},
		{RoomState(7).String(), "RoomState(7)"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("String() = %q, want %q", tt.got, tt.want)
		}
	}
}
