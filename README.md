# Application de gestion de services

## 🎯 Objectif du projet
Ce projet consiste à **refondre une application existante** simulant un petit système de gestion de services et de réservations, en appliquant les **bonnes pratiques de conception et de développement** vues en cours.

L’application permet :
- de s’identifier par **email** (sans mot de passe, session simulée),
- de consulter la **liste des services** et leurs créneaux,
- de **réserver** un créneau disponible,
- de **consulter et annuler** ses réservations,
- et pour un administrateur, d’**ajouter** ou **supprimer** des services et des créneaux.

---

## ⚙️ Choix technologique : Golang

### Pourquoi Go ?
Le langage **Go** est particulièrement adapté à ce type de refactoring pour plusieurs raisons :

- 🧩 **Simplicité et lisibilité** : la syntaxe claire de Go favorise la mise en place de bonnes pratiques et la lisibilité du code.
- ⚙️ **Conception modulaire native** : la gestion des packages (`internal/`, `cmd/`, etc.) permet de séparer facilement les couches (HTTP, logique métier, données).
- 🚀 **Exécution rapide** : Go compile en un **binaire unique**, idéal pour un monolithe léger et performant.
- 🧱 **Architecture naturelle en couches** : la structuration par packages s’intègre parfaitement à un modèle **monolithique modulaire**.
- 🧪 **Outils intégrés** : `go fmt`, `go vet`, `go test`, `golangci-lint` permettent d’assurer la **qualité du code** sans dépendances externes lourdes.
- 💡 **Simplicité de déploiement** : pas besoin de serveur d’application externe — Go dispose de sa propre librairie HTTP.

En somme, Go favorise un **code propre, rapide et bien structuré**, ce qui correspond parfaitement à l’objectif du TP : améliorer la **qualité et la structure** d’une application sans en complexifier le fonctionnement.

---

## 🧱 Architecture choisie : Monolithique modulaire

### 🧩 Type d’architecture

Le projet adopte une **architecture monolithique modulaire**, inspirée du modèle Clean Architecture.

- **Monolithique** → tout le code (API, logique métier, stockage JSON) est réuni dans une seule application Go.  
- **Modulaire** → les différentes couches (présentation, métier, données) sont clairement séparées et découplées.

### 💬 Pourquoi ce choix ?

Ce type d’architecture est le plus adapté :
- à la **simplicité du projet**, qui ne justifie pas la complexité des microservices ;
- à la **philosophie de Go**, conçu pour des binaires uniques, performants et faciles à déployer ;
- à la **lisibilité et testabilité** : chaque couche a une responsabilité claire et peut être testée indépendamment.

En résumé, cette approche permet un code **propre, maintenable et évolutif**, tout en restant **léger et rapide à mettre en œuvre**.

---

## 📁 Arborescence simplifiée

```text
.
├─ cmd/
│  └─ api/
│     └─ main.go
├─ data/
├─ internal/
├─ web/

---

## 🏃 Lancer le serveur backend

Dans le dossier du projet :

```bash
go run ./cmd/api
```

```
Server listening on :8080
```

L'API REST tourne sur :
👉 http://localhost:8080/

---

## 🌐 Accéder au frontend

Ouvrir le navigateur et aller sur :

👉 http://localhost:8080/


---


## 🧹 Vider la pseudo-base JSON (réinitialiser l'app)

Efface les fichiers :

- `data/services.json`
- `data/slots.json`
- `data/reservations.json`

Et remetre pour chacun :

```
[]
```

Puis relance le serveur.

---