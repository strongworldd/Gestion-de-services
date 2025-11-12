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

type jsonDB struct {
	Services     []services.Service     `json:"services"`
	Slots        []services.Slot        `json:"slots"`
	Reservations []services.Reservation `json:"reservations"`
}

// Génération ID simple: prefix + timestamp
func newID(prefix string) services.ID {
	return services.ID(fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano()))
}

type JSONStore struct {
	mu     sync.Mutex
	root   string
	db     jsonDB
	loaded bool
}

func NewJSONStore(root string) (*JSONStore, error) {
	js := &JSONStore{root: root}
	if err := js.load(); err != nil {
		return nil, err
	}
	return js, nil
}

func (s *JSONStore) dataPath(name string) string {
	return filepath.Join(s.root, name)
}

func (s *JSONStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ensure := func(name string) error {
		p := s.dataPath(name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return os.WriteFile(p, []byte("[]"), 0644)
		}
		return nil
	}
	if err := ensure("services.json"); err != nil { return err }
	if err := ensure("slots.json"); err != nil { return err }
	if err := ensure("reservations.json"); err != nil { return err }

	readJSON := func(name string, v any) error {
		b, err := os.ReadFile(s.dataPath(name))
		if err != nil { return err }
		if len(b) == 0 { b = []byte("[]") }
		return json.Unmarshal(b, v)
	}

	if err := readJSON("services.json", &s.db.Services); err != nil { return err }
	if err := readJSON("slots.json", &s.db.Slots); err != nil { return err }
	if err := readJSON("reservations.json", &s.db.Reservations); err != nil { return err }

	s.loaded = true
	return nil
}

func (s *JSONStore) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeJSON := func(name string, v any) error {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil { return err }
		return os.WriteFile(s.dataPath(name), b, 0644)
	}

	if err := writeJSON("services.json", s.db.Services); err != nil { return err }
	if err := writeJSON("slots.json", s.db.Slots); err != nil { return err }
	if err := writeJSON("reservations.json", s.db.Reservations); err != nil { return err }
	return nil
}

// ---------- Services ----------

func (s *JSONStore) ListServices() ([]services.Service, error) {
	if !s.loaded { if err := s.load(); err != nil { return nil, err } }
	return append([]services.Service(nil), s.db.Services...), nil
}

func (s *JSONStore) CreateService(svc services.Service) (services.Service, error) {
	if svc.ID == "" {
		svc.ID = newID("svc")
	}
	s.db.Services = append(s.db.Services, svc)
	if err := s.save(); err != nil { return services.Service{}, err }
	return svc, nil
}

// ---------- Slots ----------

func (s *JSONStore) AddSlot(slot services.Slot) (services.Slot, error) {
	if slot.ID == "" {
		slot.ID = newID("slt")
	}
	s.db.Slots = append(s.db.Slots, slot)
	if err := s.save(); err != nil { return services.Slot{}, err }
	return slot, nil
}

func (s *JSONStore) ListSlotsByService(serviceID services.ID) ([]services.Slot, error) {
	var out []services.Slot
	for _, sl := range s.db.Slots {
		if sl.ServiceID == serviceID {
			out = append(out, sl)
		}
	}
	return out, nil
}

func (s *JSONStore) GetSlot(slotID services.ID) (services.Slot, error) {
	for _, sl := range s.db.Slots {
		if sl.ID == slotID {
			return sl, nil
		}
	}
	return services.Slot{}, errors.New("slot not found")
}

// ---------- Reservations ----------

func (s *JSONStore) CreateReservation(r services.Reservation) (services.Reservation, error) {
	if r.ID == "" {
		r.ID = newID("res")
	}
	s.db.Reservations = append(s.db.Reservations, r)
	if err := s.save(); err != nil { return services.Reservation{}, err }
	return r, nil
}

func (s *JSONStore) ListReservationsByEmail(email string) ([]services.Reservation, error) {
	var out []services.Reservation
	for _, r := range s.db.Reservations {
		if r.UserEmail == email {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *JSONStore) ListReservationsBySlot(slotID services.ID) ([]services.Reservation, error) {
	var out []services.Reservation
	for _, r := range s.db.Reservations {
		if r.SlotID == slotID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *JSONStore) GetReservation(resID services.ID) (services.Reservation, error) {
	for _, r := range s.db.Reservations {
		if r.ID == resID {
			return r, nil
		}
	}
	return services.Reservation{}, errors.New("reservation not found")
}

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
	s.db.Reservations = append(s.db.Reservations[:idx], s.db.Reservations[idx+1:]...)
	return s.save()
}