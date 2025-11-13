package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"gestionsvc/internal/services"
)

//
// ---------- Structure du serveur HTTP ----------
//

// Server regroupe :
// - un ServeMux pour enregistrer les routes HTTP,
// - un BookingService qui contient la logique métier.
type Server struct {
	Mux     *http.ServeMux
	Booking *services.BookingService
}

// NewServer configure les routes de l'API et retourne un Server prêt à être utilisé.
func NewServer(b *services.BookingService) *Server {
	s := &Server{
		Mux:     http.NewServeMux(),
		Booking: b,
	}

	// Auth simulée
	s.Mux.HandleFunc("/auth/login", s.login)

	// Services
	s.Mux.HandleFunc("/services", s.listServices)      // GET /services
	s.Mux.HandleFunc("/services/", s.serviceSubroutes) // GET /services/:id/slots

	// Administration
	s.Mux.HandleFunc("/admin/services", s.adminCreateService)      // POST /admin/services
	s.Mux.HandleFunc("/admin/services/", s.adminServiceSubroutes)  // POST /admin/services/:id/slots

	// Réservations
	s.Mux.HandleFunc("/reservations", s.reservationsRoot) // POST /reservations
	s.Mux.HandleFunc("/reservations/", s.reservationsSub) // GET /reservations/me, DELETE /reservations/:id

	return s
}

//
// ---------- Helpers génériques JSON / Auth ----------
//

// writeJSON sérialise une valeur en JSON et écrit le code HTTP spécifié.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// readJSON lit le corps de la requête et le désérialise dans v.
func readJSON(r *http.Request, v any) error {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if len(b) == 0 {
		return nil
	}

	return json.Unmarshal(b, v)
}

// currentEmail récupère l'email courant depuis l'en-tête HTTP "X-User-Email".
func currentEmail(r *http.Request) string {
	return r.Header.Get("X-User-Email")
}

// isAdmin vérifie si l'email correspond à l'administrateur.
func isAdmin(email string) bool {
	return email == "admin@example.com"
}

//
// ---------- Auth simulée ----------
//

// POST /auth/login
//
// Le front envoie un JSON { "email": "..." }.
// Ici on ne gère pas de session réelle : on renvoie juste l'email,
// et le front le stocke côté navigateur (localStorage).
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var in struct {
		Email string `json:"email"`
	}

	_ = readJSON(r, &in)

	if in.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email required"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"email": in.Email})
}

//
// ---------- Services ----------
//

// GET /services
//
// Retourne la liste des services disponibles.
func (s *Server) listServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	list, err := s.Booking.ListServices()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, list)
}

// GET /services/:id/slots
//
// Gère les sous-routes de /services/.
// Exemple d'URL attendue : /services/svc_123/slots
func (s *Server) serviceSubroutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// On attend : [ "services", ":id", "slots" ]
	if len(parts) == 3 && parts[0] == "services" && parts[2] == "slots" {
		svcID := services.ID(parts[1])

		slots, err := s.Booking.ListSlotsByService(svcID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, slots)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

//
// ---------- Administration ----------
//

// POST /admin/services
//
// Crée un nouveau service.
//
// Nécessite l'en-tête X-User-Email = admin@example.com
func (s *Server) adminCreateService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if !isAdmin(currentEmail(r)) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin only"})
		return
	}

	var in struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Duration    int    `json:"duration"`
	}

	if err := readJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad json"})
		return
	}

	svc, err := s.Booking.CreateService(in.Name, in.Description, in.Duration)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, svc)
}

// POST /admin/services/:id/slots
//
// Ajoute un créneau à un service existant.
// Body JSON : { "datetime": "...", "capacity": 1 }
//
// Toujours réservé à l'admin.
func (s *Server) adminServiceSubroutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if !isAdmin(currentEmail(r)) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin only"})
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// On attend : [ "admin", "services", ":id", "slots" ]
	if len(parts) == 4 && parts[0] == "admin" && parts[1] == "services" && parts[3] == "slots" {
		svcID := services.ID(parts[2])

		var in struct {
			Datetime string `json:"datetime"`
			Capacity int    `json:"capacity"`
		}

		if err := readJSON(r, &in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad json"})
			return
		}

		slot, err := s.Booking.AddSlot(svcID, in.Datetime, in.Capacity)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, slot)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

//
// ---------- Réservations ----------
//

// POST /reservations
//
// Crée une réservation pour l'utilisateur courant (X-User-Email).
//
// Body JSON : { "slotId": "slt_123" }
func (s *Server) reservationsRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		em := currentEmail(r)

		var in struct {
			SlotID services.ID `json:"slotId"`
		}

		if err := readJSON(r, &in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad json"})
			return
		}

		res, err := s.Booking.Book(in.SlotID, em)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, res)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// GET /reservations/me
// DELETE /reservations/:id
//
// - GET /reservations/me : toutes les réservations de l'utilisateur courant.
// - DELETE /reservations/res_123 : annule une réservation (si encore valable).
func (s *Server) reservationsSub(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "reservations" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// /reservations/me
	if len(parts) == 2 && parts[1] == "me" && r.Method == http.MethodGet {
		em := currentEmail(r)

		list, err := s.Booking.MyReservations(em)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, list)
		return
	}

	// /reservations/:id
	if len(parts) == 2 && r.Method == http.MethodDelete {
		em := currentEmail(r)
		id := services.ID(parts[1])

		if err := s.Booking.Cancel(id, em); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
		return
	}

	// Méthode non supportée sur cette sous-route
	w.WriteHeader(http.StatusMethodNotAllowed)
}