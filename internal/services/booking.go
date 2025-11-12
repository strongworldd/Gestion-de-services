package services

import (
	"errors"
	"time"
)

// ---------- Modèles domaine ----------

type ID string

type Service struct {
	ID          ID     `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Duration    int    `json:"duration,omitempty"` // minutes (optionnel)
}

type Slot struct {
	ID        ID       `json:"id"`
	ServiceID ID       `json:"serviceId"`
	Datetime  time.Time `json:"datetime"`
	Capacity  int      `json:"capacity"`
}

type Reservation struct {
	ID        ID       `json:"id"`
	SlotID    ID       `json:"slotId"`
	UserEmail string   `json:"userEmail"`
	CreatedAt time.Time `json:"createdAt"`
}

// ---------- Contrat repository (persistance) ----------

type Repository interface {
	// Services
	ListServices() ([]Service, error)
	CreateService(s Service) (Service, error)

	// Slots
	AddSlot(slot Slot) (Slot, error)
	ListSlotsByService(serviceID ID) ([]Slot, error)
	GetSlot(slotID ID) (Slot, error)

	// Reservations
	CreateReservation(r Reservation) (Reservation, error)
	ListReservationsByEmail(email string) ([]Reservation, error)
	ListReservationsBySlot(slotID ID) ([]Reservation, error)
	DeleteReservation(resID ID) error
	GetReservation(resID ID) (Reservation, error)
}

// ---------- Service métier ----------

type BookingService struct {
	repo Repository
	now  func() time.Time
}

func NewBookingService(r Repository) *BookingService {
	return &BookingService{
		repo: r,
		now:  time.Now,
	}
}

// -- Admin --

func (b *BookingService) CreateService(name, desc string, duration int) (Service, error) {
	if name == "" {
		return Service{}, errors.New("name required")
	}
	return b.repo.CreateService(Service{
		Name:        name,
		Description: desc,
		Duration:    duration,
	})
}

func (b *BookingService) AddSlot(serviceID ID, isoDatetime string, capacity int) (Slot, error) {
	if capacity <= 0 {
		capacity = 1
	}
	t, err := time.Parse(time.RFC3339, isoDatetime)
	if err != nil {
		return Slot{}, errors.New("invalid datetime (use RFC3339)")
	}
	slot := Slot{
		ServiceID: serviceID,
		Datetime:  t,
		Capacity:  capacity,
	}
	return b.repo.AddSlot(slot)
}

// -- Public --

func (b *BookingService) ListServices() ([]Service, error) {
	return b.repo.ListServices()
}

func (b *BookingService) ListSlotsByService(svcID ID) ([]Slot, error) {
	return b.repo.ListSlotsByService(svcID)
}

func (b *BookingService) Book(slotID ID, userEmail string) (Reservation, error) {
	if userEmail == "" {
		return Reservation{}, errors.New("missing user email")
	}
	slot, err := b.repo.GetSlot(slotID)
	if err != nil {
		return Reservation{}, errors.New("slot not found")
	}

	// Règle 1: pas de double booking exact pour le même user + slot
	existing, _ := b.repo.ListReservationsBySlot(slotID)
	for _, r := range existing {
		if r.UserEmail == userEmail {
			return Reservation{}, errors.New("already booked this slot")
		}
	}

	// Règle 2: capacité simple
	if len(existing) >= slot.Capacity {
		return Reservation{}, errors.New("slot is full")
	}

	// OK → créer réservation
	return b.repo.CreateReservation(Reservation{
		SlotID:    slotID,
		UserEmail: userEmail,
		CreatedAt: b.now(),
	})
}

func (b *BookingService) MyReservations(userEmail string) ([]Reservation, error) {
	return b.repo.ListReservationsByEmail(userEmail)
}

func (b *BookingService) Cancel(resID ID, userEmail string) error {
	res, err := b.repo.GetReservation(resID)
	if err != nil {
		return errors.New("reservation not found")
	}
	// Option : autoriser l’annulation seulement pour l’owner
	if res.UserEmail != userEmail {
		return errors.New("not your reservation")
	}
	// Option (si vous voulez) : annulation uniquement pour le futur
	slot, err := b.repo.GetSlot(res.SlotID)
	if err == nil && !slot.Datetime.After(b.now()) {
		return errors.New("cannot cancel past reservations")
	}

	return b.repo.DeleteReservation(resID)
}