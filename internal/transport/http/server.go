package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"gestionsvc/internal/services"
)

type Server struct {
	Mux     *http.ServeMux
	Booking *services.BookingService
}

func NewServer(b *services.BookingService) *Server {
	s := &Server{
		Mux:     http.NewServeMux(),
		Booking: b,
	}
	// Auth simulée
	s.Mux.HandleFunc("/auth/login", s.login)

	// Services
	s.Mux.HandleFunc("/services", s.listServices)               // GET
	s.Mux.HandleFunc("/services/", s.serviceSubroutes)          // GET /services/:id/slots

	// Admin
	s.Mux.HandleFunc("/admin/services", s.adminCreateService)               // POST
	s.Mux.HandleFunc("/admin/services/", s.adminServiceSubroutes)          // POST /admin/services/:id/slots

	// Reservations
	s.Mux.HandleFunc("/reservations", s.reservationsRoot)       // POST, GET? (non)
	s.Mux.HandleFunc("/reservations/", s.reservationsSub)       // GET /me  | DELETE /:id

	return s
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v any) error {
	b, err := io.ReadAll(r.Body)
	if err != nil { return err }
	defer r.Body.Close()
	if len(b) == 0 { return nil }
	return json.Unmarshal(b, v)
}

func currentEmail(r *http.Request) string {
	return r.Header.Get("X-User-Email")
}

func isAdmin(email string) bool {
	return email == "admin@example.com"
}

// ---------- Auth simulée ----------

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed); return
	}
	var in struct{ Email string `json:"email"` }
	_ = readJSON(r, &in)
	if in.Email == "" {
		writeJSON(w, 400, map[string]string{"error":"email required"})
		return
	}
	writeJSON(w, 200, map[string]string{"email": in.Email})
}

// ---------- Services ----------

func (s *Server) listServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405); return
	}
	list, err := s.Booking.ListServices()
	if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
	writeJSON(w, 200, list)
}

// /services/:id/slots
func (s *Server) serviceSubroutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405); return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// expect: [services, :id, slots]
	if len(parts) == 3 && parts[0] == "services" && parts[2] == "slots" {
		svcID := services.ID(parts[1])
		slots, err := s.Booking.ListSlotsByService(svcID)
		if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
		writeJSON(w, 200, slots)
		return
	}
	w.WriteHeader(404)
}

// ---------- Admin ----------

// POST /admin/services
func (s *Server) adminCreateService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(405); return }
	if !isAdmin(currentEmail(r)) {
		writeJSON(w, 403, map[string]string{"error":"admin only"})
		return
	}
	var in struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Duration    int    `json:"duration"`
	}
	if err := readJSON(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error":"bad json"}); return
	}
	svc, err := s.Booking.CreateService(in.Name, in.Description, in.Duration)
	if err != nil { writeJSON(w, 400, map[string]string{"error": err.Error()}); return }
	writeJSON(w, 200, svc)
}

// POST /admin/services/:id/slots
func (s *Server) adminServiceSubroutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(405); return }
	if !isAdmin(currentEmail(r)) {
		writeJSON(w, 403, map[string]string{"error":"admin only"}); return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 4 && parts[0] == "admin" && parts[1] == "services" && parts[3] == "slots" {
		svcID := services.ID(parts[2])
		var in struct {
			Datetime string `json:"datetime"`
			Capacity int    `json:"capacity"`
		}
		if err := readJSON(r, &in); err != nil {
			writeJSON(w, 400, map[string]string{"error":"bad json"}); return
		}
		slot, err := s.Booking.AddSlot(svcID, in.Datetime, in.Capacity)
		if err != nil { writeJSON(w, 400, map[string]string{"error": err.Error()}); return }
		writeJSON(w, 200, slot); return
	}
	w.WriteHeader(404)
}

// ---------- Reservations ----------

// POST /reservations
func (s *Server) reservationsRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		em := currentEmail(r)
		var in struct{ SlotID services.ID `json:"slotId"` }
		if err := readJSON(r, &in); err != nil {
			writeJSON(w, 400, map[string]string{"error":"bad json"}); return
		}
		res, err := s.Booking.Book(in.SlotID, em)
		if err != nil { writeJSON(w, 400, map[string]string{"error": err.Error()}); return }
		writeJSON(w, 200, res)
	default:
		w.WriteHeader(405)
	}
}

// GET /reservations/me
// DELETE /reservations/:id
func (s *Server) reservationsSub(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "reservations" {
		w.WriteHeader(404); return
	}

	// /reservations/me
	if len(parts) == 2 && parts[1] == "me" && r.Method == http.MethodGet {
		em := currentEmail(r)
		list, err := s.Booking.MyReservations(em)
		if err != nil { writeJSON(w, 500, map[string]string{"error": err.Error()}); return }
		writeJSON(w, 200, list); return
	}

	// /reservations/:id
	if len(parts) == 2 && r.Method == http.MethodDelete {
		em := currentEmail(r)
		id := services.ID(parts[1])
		if err := s.Booking.Cancel(id, em); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()}); return
		}
		writeJSON(w, 200, map[string]string{"status":"deleted"}); return
	}

	w.WriteHeader(405)
}