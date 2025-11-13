package services

import (
	"errors"
	"time"
)

//
// ---------- Modèles du domaine ----------
//

// ID représente un identifiant unique pour les entités (services, slots, réservations)
type ID string

// Service correspond à un type de prestation (ex : Coiffure, Massage, etc.)
type Service struct {
	ID          ID     `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Duration    int    `json:"duration,omitempty"` // Durée en minutes (facultatif)
}

// Slot = créneau horaire disponible pour un service donné
type Slot struct {
	ID        ID        `json:"id"`
	ServiceID ID        `json:"serviceId"`
	Datetime  time.Time `json:"datetime"`
	Capacity  int       `json:"capacity"`
}

// Reservation = une réservation effectuée par un utilisateur sur un slot
type Reservation struct {
	ID        ID        `json:"id"`
	SlotID    ID        `json:"slotId"`
	UserEmail string    `json:"userEmail"`
	CreatedAt time.Time `json:"createdAt"`
}

//
// ---------- Interface du Repository (contrat de persistance) ----------
//

// Repository définit les méthodes nécessaires pour stocker et lire les données.
// La couche métier (BookingService) utilise cette interface sans savoir si les
// données sont stockées en JSON, SQL, mémoire, etc.
type Repository interface {
	// Services
	ListServices() ([]Service, error)
	CreateService(s Service) (Service, error)

	// Slots
	AddSlot(slot Slot) (Slot, error)
	ListSlotsByService(serviceID ID) ([]Slot, error)
	GetSlot(slotID ID) (Slot, error)

	// Réservations
	CreateReservation(r Reservation) (Reservation, error)
	ListReservationsByEmail(email string) ([]Reservation, error)
	ListReservationsBySlot(slotID ID) ([]Reservation, error)
	GetReservation(resID ID) (Reservation, error)
	DeleteReservation(resID ID) error
}

//
// ---------- Service métier (logique principale) ----------
//

// BookingService contient la logique de réservation.
// Il utilise un Repository pour lire/écrire les données.
type BookingService struct {
	repo Repository
	now  func() time.Time
}

// NewBookingService instancie un nouveau service métier
func NewBookingService(r Repository) *BookingService {
	return &BookingService{
		repo: r,
		now:  time.Now, // permet de mocker la date en tests
	}
}

//
// ---------- Logique Admin ----------
//

// CreateService permet de créer un service (admin uniquement)
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

// AddSlot crée un créneau horaire pour un service donné.
// Le datetime doit être au format RFC3339.
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

//
// ---------- Logique publique ----------
//

// ListServices retourne tous les services disponibles
func (b *BookingService) ListServices() ([]Service, error) {
	return b.repo.ListServices()
}

// ListSlotsByService retourne les créneaux d'un service donné
func (b *BookingService) ListSlotsByService(svcID ID) ([]Slot, error) {
	return b.repo.ListSlotsByService(svcID)
}

// Book tente de réserver un créneau
func (b *BookingService) Book(slotID ID, userEmail string) (Reservation, error) {
	if userEmail == "" {
		return Reservation{}, errors.New("missing user email")
	}

	// Vérifier que le créneau existe
	slot, err := b.repo.GetSlot(slotID)
	if err != nil {
		return Reservation{}, errors.New("slot not found")
	}

	// 1) L'utilisateur ne peut pas réserver deux fois le même slot
	existing, _ := b.repo.ListReservationsBySlot(slotID)
	for _, r := range existing {
		if r.UserEmail == userEmail {
			return Reservation{}, errors.New("already booked this slot")
		}
	}

	// 2) Vérifier la capacité maximale
	if len(existing) >= slot.Capacity {
		return Reservation{}, errors.New("slot is full")
	}

	// OK → création de la réservation
	return b.repo.CreateReservation(Reservation{
		SlotID:    slotID,
		UserEmail: userEmail,
		CreatedAt: b.now(),
	})
}

// MyReservations retourne les réservations d'un utilisateur
func (b *BookingService) MyReservations(userEmail string) ([]Reservation, error) {
	return b.repo.ListReservationsByEmail(userEmail)
}

// Cancel annule une réservation si :
// - elle existe
// - elle appartient à l'utilisateur
// - elle concerne un créneau futur
func (b *BookingService) Cancel(resID ID, userEmail string) error {
	res, err := b.repo.GetReservation(resID)
	if err != nil {
		return errors.New("reservation not found")
	}

	// Vérifier que c’est bien la réservation de cet utilisateur
	if res.UserEmail != userEmail {
		return errors.New("not your reservation")
	}

	// Vérifier que le créneau n'est pas passé
	slot, err := b.repo.GetSlot(res.SlotID)
	if err == nil && !slot.Datetime.After(b.now()) {
		return errors.New("cannot cancel past reservations")
	}

	return b.repo.DeleteReservation(resID)
}