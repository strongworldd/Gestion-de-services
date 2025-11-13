package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gestionsvc/internal/services"
)

//
// ---------- Structure interne du "JSON-DB" ----------
//

// jsonDB représente la structure complète des données
// telles qu'elles sont stockées dans les fichiers JSON.
//
// Elle agit comme une "base de données" chargée en mémoire,
// et sera régulièrement sauvegardée sur disque.
type jsonDB struct {
	Services     []services.Service     `json:"services"`
	Slots        []services.Slot        `json:"slots"`
	Reservations []services.Reservation `json:"reservations"`
}

//
// ---------- Helpers ----------
//

// newID génère un identifiant unique simple, basé sur
// un préfixe + un timestamp haute résolution.
// Exemple : "svc_1731965329823345000"
func newID(prefix string) services.ID {
	return services.ID(fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano()))
}

//
// ---------- JSONStore : implémentation du Repository ----------
//

// JSONStore gère la lecture/écriture des fichiers JSON.
// • root = dossier local contenant les fichiers services.json, slots.json...
// • db = copie en mémoire des données
// • mu = évite les accès concurrents
type JSONStore struct {
	mu     sync.Mutex
	root   string
	db     jsonDB
	loaded bool
}

// NewJSONStore crée un store et charge immédiatement les fichiers JSON.
func NewJSONStore(root string) (*JSONStore, error) {
	js := &JSONStore{root: root}
	if err := js.load(); err != nil {
		return nil, err
	}
	return js, nil
}

// dataPath renvoie le chemin complet d'un fichier du store.
func (s *JSONStore) dataPath(name string) string {
	return filepath.Join(s.root, name)
}

//
// ---------- Chargement / Sauvegarde ----------
//

// load lit les fichiers JSON du disque et remplit s.db.
//
// Si un fichier n'existe pas encore, il est créé automatiquement
// avec un tableau vide "[]".
func (s *JSONStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Assure que les fichiers existent
	ensure := func(name string) error {
		p := s.dataPath(name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			// Fichier manquant → on crée un fichier vide
			return os.WriteFile(p, []byte("[]"), 0o644)
		}
		return nil
	}

	if err := ensure("services.json"); err != nil {
		return err
	}
	if err := ensure("slots.json"); err != nil {
		return err
	}
	if err := ensure("reservations.json"); err != nil {
		return err
	}

	// Fonction utilitaire pour lire un fichier JSON
	readJSON := func(name string, v any) error {
		b, err := os.ReadFile(s.dataPath(name))
		if err != nil {
			return err
		}
		if len(b) == 0 {
			b = []byte("[]")
		}
		return json.Unmarshal(b, v)
	}

	// Lecture des données
	if err := readJSON("services.json", &s.db.Services); err != nil {
		return err
	}
	if err := readJSON("slots.json", &s.db.Slots); err != nil {
		return err
	}
	if err := readJSON("reservations.json", &s.db.Reservations); err != nil {
		return err
	}

	s.loaded = true
	return nil
}

// save écrit s.db dans les fichiers JSON.
//
// C'est ici que la persistance est réellement effectuée à chaque modification.
func (s *JSONStore) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeJSON := func(name string, v any) error {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(s.dataPath(name), b, 0o644)
	}

	if err := writeJSON("services.json", s.db.Services); err != nil {
		return err
	}
	if err := writeJSON("slots.json", s.db.Slots); err != nil {
		return err
	}
	if err := writeJSON("reservations.json", s.db.Reservations); err != nil {
		return err
	}

	return nil
}

//
// ---------- Services ----------
//

// ListServices renvoie la liste des services.
// On renvoie une copie pour éviter que l’appelant ne modifie directement le slice interne.
func (s *JSONStore) ListServices() ([]services.Service, error) {
	if !s.loaded {
		if err := s.load(); err != nil {
			return nil, err
		}
	}
	return append([]services.Service(nil), s.db.Services...), nil
}

// CreateService ajoute un nouveau service dans le store.
func (s *JSONStore) CreateService(svc services.Service) (services.Service, error) {
	if svc.ID == "" {
		svc.ID = newID("svc")
	}

	s.db.Services = append(s.db.Services, svc)

	if err := s.save(); err != nil {
		return services.Service{}, err
	}

	return svc, nil
}

//
// ---------- Slots ----------
//

// AddSlot ajoute un créneau horaire (slot) à la liste.
func (s *JSONStore) AddSlot(slot services.Slot) (services.Slot, error) {
	if slot.ID == "" {
		slot.ID = newID("slt")
	}

	s.db.Slots = append(s.db.Slots, slot)

	if err := s.save(); err != nil {
		return services.Slot{}, err
	}

	return slot, nil
}

// ListSlotsByService retourne tous les créneaux liés à un service donné.
func (s *JSONStore) ListSlotsByService(serviceID services.ID) ([]services.Slot, error) {
	var out []services.Slot
	for _, sl := range s.db.Slots {
		if sl.ServiceID == serviceID {
			out = append(out, sl)
		}
	}
	return out, nil
}

// GetSlot retourne un slot selon son ID.
func (s *JSONStore) GetSlot(slotID services.ID) (services.Slot, error) {
	for _, sl := range s.db.Slots {
		if sl.ID == slotID {
			return sl, nil
		}
	}
	return services.Slot{}, errors.New("slot not found")
}

//
// ---------- Reservations ----------
//

// CreateReservation enregistre une réservation.
func (s *JSONStore) CreateReservation(r services.Reservation) (services.Reservation, error) {
	if r.ID == "" {
		r.ID = newID("res")
	}

	s.db.Reservations = append(s.db.Reservations, r)

	if err := s.save(); err != nil {
		return services.Reservation{}, err
	}

	return r, nil
}

// ListReservationsByEmail recherche toutes les réservations d’un utilisateur.
func (s *JSONStore) ListReservationsByEmail(email string) ([]services.Reservation, error) {
	var out []services.Reservation
	for _, r := range s.db.Reservations {
		if r.UserEmail == email {
			out = append(out, r)
		}
	}
	return out, nil
}

// ListReservationsBySlot retourne les réservations d’un créneau donné.
func (s *JSONStore) ListReservationsBySlot(slotID services.ID) ([]services.Reservation, error) {
	var out []services.Reservation
	for _, r := range s.db.Reservations {
		if r.SlotID == slotID {
			out = append(out, r)
		}
	}
	return out, nil
}

// GetReservation récupère une réservation par ID.
func (s *JSONStore) GetReservation(resID services.ID) (services.Reservation, error) {
	for _, r := range s.db.Reservations {
		if r.ID == resID {
			return r, nil
		}
	}
	return services.Reservation{}, errors.New("reservation not found")
}

// DeleteReservation supprime une réservation si elle existe.
func (s *JSONStore) DeleteReservation(resID services.ID) error {
	idx := -1

	for i, r := range s.db.Reservations {
		if r.ID == resID {
			idx = i
			break
		}
	}

	if idx < 0 {
		return errors.New("reservation not found")
	}

	// Suppression propre du slice
	s.db.Reservations = append(
		s.db.Reservations[:idx],
		s.db.Reservations[idx+1:]...,
	)

	return s.save()
}