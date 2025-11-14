# 📘 Documentation Backend – Application de Gestion de Services

Ce document explique clairement le fonctionnement du **backend en Go**, son architecture, et comment les différentes couches communiquent entre elles.

---

# 🧩 Architecture générale du backend

Le backend utilise une architecture propre et modulaire :

```
/internal
 ├── transport/http     → API REST (server.go)
 ├── services           → Logique métier (booking.go)
 └── repository         → Persistance des données (jsonstore.go)
main.go                 → Assemble tout et lance le serveur
```

---

# ⚙️ 1. API REST — `server.go`

📍 *Dossier :* `internal/transport/http`

`server.go` gère **toutes les requêtes HTTP** venant du front (JavaScript).

### Rôle :

- Recevoir les données envoyées par `fetch()` en JSON.
- Appeler la logique métier (BookingService).
- Retourner une réponse JSON au front.

### Exemples de routes :

| Méthode | Route                        | Description |
|--------|------------------------------|-------------|
| GET    | `/services`                  | Liste des services |
| GET    | `/services/:id/slots`        | Slots d’un service |
| POST   | `/auth/login`                | Connexion |
| POST   | `/reservations`              | Réserver un slot |
| GET    | `/reservations/me`           | Voir ses réservations |
| DELETE | `/reservations/:id`          | Annuler une réservation |
| POST   | `/admin/services`            | Créer un service |
| POST   | `/admin/services/:id/slots`  | Ajouter un slot |

---

# 🧠 2. Logique métier — `booking.go`

📍 *Dossier :* `internal/services`

C’est le **cerveau** de l’application.

Il gère toutes les règles métier : capacité, doublons, annulations, format de date…

### Structure principale :

```go
type BookingService struct {
    repo Repository
    now  func() time.Time
}
```

### Constructeur :

```go
func NewBookingService(r Repository) *BookingService
```

Il crée l’instance utilisée par le serveur HTTP.

---

# 💾 3. Persistance — `jsonstore.go`

📍 *Dossier :* `internal/repository`

Ce fichier stocke **physiquement** les données dans des fichiers `.json`.

`booking.go` ne sait pas comment les données sont stockées : il utilise seulement l’interface `Repository`.

On pourrait remplacer jsonstore par `sqlstore.go` sans modifier le reste du projet.

---

# 🚀 4. main.go — Point d’entrée

- Initialise le repository
- Initialise BookingService
- Crée le serveur HTTP
- Sert les fichiers du front (`/web`)
- Lance l’application sur `localhost:8080`

---

# 🎯 Résumé

| Couche | Rôle |
|-------|------|
| **server.go** | API REST : reçoit / répond en JSON |
| **booking.go** | Logique métier |
| **jsonstore.go** | Stockage des données |
| **main.go** | Assemble et démarre le backend |

---

Backend prêt, clair et modulaire !
