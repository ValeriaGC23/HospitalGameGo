// This file runs the web interface (go run . -web): a small HTTP server that
// serves the page in web/index.html and a JSON API over the Hospital. It
// only uses the public API of package hospital, so the logic stays in that
// package and this file is only the view.

package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"JuegoHospitalGo/hospital"
)

//go:embed web
var webFiles embed.FS

// server holds the hospital shown in the page and the state of the game.
// Its own mutex protects these fields; the hospital protects itself.
//
// The game has two switches:
//   - running (Start/Pause): while it is on, every patient has a goroutine
//     that makes them fall asleep at random times, one goroutine moves
//     hallway patients to free rooms and another wakes anyone asleep longer
//     than maxSleep.
//   - auto (Auto simulation): with it on, the patients also wake up by
//     themselves. With it off (manual mode) the player must wake them up.
type server struct {
	start time.Time

	mu       sync.Mutex
	h        *hospital.Hospital // replaced on restart
	auto     bool
	runStop  chan struct{} // nil while the game is paused
	autoStop chan struct{} // nil while the patient goroutines are not running
}

// roster lists every patient the game can use: the 5 demo patients first,
// then 5 more. The player chooses how many of them to admit.
var roster = append(append([]patientData(nil), demoPatients...),
	patientData{"P-006", "Noah Kim", 33, hospital.Moderate},
	patientData{"P-007", "Fatima Haddad", 52, hospital.Severe},
	patientData{"P-008", "Ana Silva", 24, hospital.Mild},
	patientData{"P-009", "Omar Diallo", 45, hospital.Moderate},
	patientData{"P-010", "Mei Chen", 71, hospital.Severe},
)

const (
	minPatients     = 3
	defaultPatients = 5
)

// newGameHospital builds the hospital of the game: the scenario staff and
// rooms, and the first n patients of the roster (n is kept in range).
func newGameHospital(n int) (*hospital.Hospital, error) {
	n = max(minPatients, min(n, len(roster)))
	h, err := hospital.NewHospital("My Hospital")
	if err != nil {
		return nil, err
	}
	c := &cast{}
	if err := hireStaffAndRooms(h, c); err != nil {
		return nil, err
	}
	if err := admitPatients(h, c, roster[:n]); err != nil {
		return nil, err
	}
	return h, nil
}

// runWeb builds the game hospital and serves the interface at addr.
func runWeb(addr string) error {
	h, err := newGameHospital(defaultPatients)
	if err != nil {
		return err
	}
	s := &server{h: h, start: time.Now(), auto: true}

	page, err := fs.Sub(webFiles, "web")
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(page)))
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/attack", s.handleAttack)
	mux.HandleFunc("/api/wake", s.handleWake)
	mux.HandleFunc("/api/assign", s.handleAssign)
	mux.HandleFunc("/api/staff", s.handleStaff)
	mux.HandleFunc("/api/run", s.handleRun)
	mux.HandleFunc("/api/auto", s.handleAuto)
	mux.HandleFunc("/api/reset", s.handleReset)

	fmt.Printf("Web interface running at http://localhost%s (Ctrl+C to stop)\n", addr)
	return http.ListenAndServe(addr, mux)
}

// request is the JSON body accepted by the POST endpoints.
type request struct {
	PatientID string `json:"patientId"`
	Location  string `json:"location"`
	StaffID   string `json:"staffId"`
	On        bool   `json:"on"`
	Patients  int    `json:"patients"`
}

// reply is the JSON answer of the POST endpoints.
type reply struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (s *server) handleState(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	h, running, auto := s.h, s.runStop != nil, s.auto
	s.mu.Unlock()
	v := toView(h.Snapshot(), auto)
	v.Running = running
	v.MaxSleep = int(maxSleep.Seconds())
	v.MaxPatients = len(roster)
	writeJSON(w, v)
}

func (s *server) handleAttack(w http.ResponseWriter, r *http.Request) {
	req, h, p, ok := s.readPatient(w, r)
	if !ok {
		return
	}
	e, err := h.HandleSleepAttack(p, req.Location)
	switch {
	case errors.Is(err, hospital.ErrNoRoomAvailable):
		writeJSON(w, reply{false, fmt.Sprintf("%s fell asleep at %s and was attended by %s, but there is no free room: they stay in the hallway.",
			p.Name(), e.Location(), e.AttendedBy().Name())})
	case err != nil:
		writeJSON(w, reply{false, err.Error()})
	default:
		writeJSON(w, reply{true, fmt.Sprintf("%s fell asleep at %s. %s took them to room %s.",
			p.Name(), e.Location(), e.AttendedBy().Name(), e.Room().ID())})
	}
}

func (s *server) handleWake(w http.ResponseWriter, r *http.Request) {
	_, h, p, ok := s.readPatient(w, r)
	if !ok {
		return
	}
	if err := h.WakePatient(p); err != nil {
		writeJSON(w, reply{false, err.Error()})
		return
	}
	writeJSON(w, reply{true, p.Name() + " woke up."})
}

func (s *server) handleAssign(w http.ResponseWriter, r *http.Request) {
	_, h, p, ok := s.readPatient(w, r)
	if !ok {
		return
	}
	room, err := h.AssignRoom(p)
	if err != nil {
		writeJSON(w, reply{false, err.Error()})
		return
	}
	writeJSON(w, reply{true, fmt.Sprintf("%s was moved to room %s.", p.Name(), room.ID())})
}

func (s *server) handleStaff(w http.ResponseWriter, r *http.Request) {
	req, ok := readRequest(w, r)
	if !ok {
		return
	}
	// The lock makes "count the doctors on duty" and "send this one off
	// duty" one step, so two clicks at once cannot leave zero doctors.
	s.mu.Lock()
	defer s.mu.Unlock()
	if !req.On && isLastDoctorOnDuty(s.h.Snapshot(), req.StaffID) {
		writeJSON(w, reply{false, "At least one doctor must stay on duty: severe patients can only be attended by a doctor."})
		return
	}
	if err := s.h.SetAttenderAvailable(req.StaffID, req.On); err != nil {
		writeJSON(w, reply{false, err.Error()})
		return
	}
	status := "off duty"
	if req.On {
		status = "on duty"
	}
	writeJSON(w, reply{true, req.StaffID + " is now " + status + "."})
}

// handleRun starts (Start/Resume) or pauses (Pause) the game.
func (s *server) handleRun(w http.ResponseWriter, r *http.Request) {
	req, ok := readRequest(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if req.On {
		s.startRun()
		writeJSON(w, reply{true, "Game running."})
		return
	}
	s.stopRun()
	writeJSON(w, reply{true, "Game paused."})
}

// handleAuto switches between auto mode (patients wake up by themselves)
// and manual mode (the player wakes them up). The patient goroutines are
// restarted so they pick up the new mode.
func (s *server) handleAuto(w http.ResponseWriter, r *http.Request) {
	req, ok := readRequest(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.auto = req.On
	if s.runStop != nil {
		s.stopPatients()
		s.startPatients()
	}
	if s.auto {
		writeJSON(w, reply{true, "Auto mode: patients fall asleep and wake up by themselves."})
		return
	}
	writeJSON(w, reply{true, "Manual mode: patients still fall asleep by themselves. Your job is to wake them up once they have spent 5 s in a room."})
}

// isLastDoctorOnDuty reports whether id is the only doctor still on duty.
func isLastDoctorOnDuty(snap hospital.Snapshot, id string) bool {
	onDuty, isDoctor := 0, false
	for _, a := range snap.Staff {
		if a.Role == "Doctor" && a.Available {
			onDuty++
			if a.ID == id {
				isDoctor = true
			}
		}
	}
	return isDoctor && onDuty == 1
}

// handleReset stops the simulation and replaces the hospital with a fresh
// one with the chosen number of patients. The simulation stays stopped, so
// the page shows the welcome screen again.
func (s *server) handleReset(w http.ResponseWriter, r *http.Request) {
	req, ok := readRequest(w, r)
	if !ok {
		return
	}
	n := req.Patients
	if n == 0 {
		n = defaultPatients
	}
	h, err := newGameHospital(n)
	if err != nil {
		writeJSON(w, reply{false, err.Error()})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopRun()
	s.h = h
	writeJSON(w, reply{true, fmt.Sprintf("Hospital restarted with %d patients.", len(h.Patients()))})
}

// startRun starts every goroutine of the game. The caller holds s.mu.
func (s *server) startRun() {
	if s.runStop != nil {
		return
	}
	s.runStop = make(chan struct{})
	go moveHallwayPatients(s.h, s.start, s.runStop)
	go wakeLongSleepers(s.h, s.start, s.runStop)
	s.startPatients()
}

// stopRun stops every goroutine of the game. The caller holds s.mu.
func (s *server) stopRun() {
	if s.runStop != nil {
		close(s.runStop)
		s.runStop = nil
	}
	s.stopPatients()
}

// startPatients launches one goroutine per patient that makes them fall
// asleep, and in auto mode also wake up. The caller holds s.mu.
func (s *server) startPatients() {
	if s.autoStop != nil {
		return
	}
	s.autoStop = make(chan struct{})
	for _, p := range s.h.Patients() {
		go simulatePatient(s.h, p, gamePace, s.auto, s.start, s.autoStop)
	}
}

// stopPatients closes their stop channel, which makes every patient
// goroutine return. The caller holds s.mu.
func (s *server) stopPatients() {
	if s.autoStop != nil {
		close(s.autoStop)
		s.autoStop = nil
	}
}

// hospital returns the current hospital (it changes on restart).
func (s *server) hospital() *hospital.Hospital {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.h
}

// readPatient reads a POST body and looks up its patient in the current
// hospital. It writes the error reply itself and returns ok=false when
// something is wrong.
func (s *server) readPatient(w http.ResponseWriter, r *http.Request) (request, *hospital.Hospital, *hospital.Patient, bool) {
	req, ok := readRequest(w, r)
	if !ok {
		return req, nil, nil, false
	}
	h := s.hospital()
	p, err := h.FindPatient(req.PatientID)
	if err != nil {
		writeJSON(w, reply{false, err.Error()})
		return req, nil, nil, false
	}
	return req, h, p, true
}

// readRequest accepts only POST requests with a JSON body.
func readRequest(w http.ResponseWriter, r *http.Request) (request, bool) {
	var req request
	if r.Method != http.MethodPost {
		http.Error(w, "use POST", http.StatusMethodNotAllowed)
		return req, false
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return req, false
	}
	return req, true
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// The view types decide how the hospital is shown in the page: states are
// sent as their String() names.

type view struct {
	Name        string         `json:"name"`
	Running     bool           `json:"running"`
	Auto        bool           `json:"auto"`
	MaxSleep    int            `json:"maxSleep"`    // seconds
	MaxPatients int            `json:"maxPatients"` // size of the roster
	Patients    []patientView  `json:"patients"`
	Staff       []staffView    `json:"staff"`
	Rooms       []roomView     `json:"rooms"`
	Episodes    []episodeView  `json:"episodes"`
	Severe      map[string]int `json:"severe"`
}

type patientView struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Level    string `json:"level"`
	State    string `json:"state"`
	Location string `json:"location"`
	RoomID   string `json:"roomId"`
}

type staffView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Available bool   `json:"available"`
	Episodes  int    `json:"episodes"`
}

type roomView struct {
	ID        string   `json:"id"`
	Capacity  int      `json:"capacity"`
	State     string   `json:"state"`
	Occupants []string `json:"occupants"`
}

type episodeView struct {
	ID          string `json:"id"`
	PatientName string `json:"patientName"`
	Location    string `json:"location"`
	AttendedBy  string `json:"attendedBy"`
	RoomID      string `json:"roomId"`
	Time        string `json:"time"`
}

// toView converts a hospital snapshot into the JSON sent to the page.
func toView(s hospital.Snapshot, auto bool) view {
	v := view{Name: s.Name, Auto: auto, Severe: s.SevereToday}
	for _, p := range s.Patients {
		v.Patients = append(v.Patients, patientView{p.ID, p.Name, p.Age, p.Level.String(), p.State.String(), p.Location, p.RoomID})
	}
	for _, a := range s.Staff {
		v.Staff = append(v.Staff, staffView{a.ID, a.Name, a.Role, a.Available, a.Episodes})
	}
	for _, r := range s.Rooms {
		v.Rooms = append(v.Rooms, roomView{r.ID, r.Capacity, r.State.String(), append([]string{}, r.Occupants...)})
	}
	for _, e := range s.Episodes {
		v.Episodes = append(v.Episodes, episodeView{e.ID, e.PatientName, e.Location, e.AttendedBy, e.RoomID, e.OccurredAt.Format("15:04:05")})
	}
	return v
}
