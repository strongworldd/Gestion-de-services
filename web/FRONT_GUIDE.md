# Application de gestion de services – Guide d’utilisation du front

## 🎯 Objectif
Ce document explique le fonctionnement de la partie **front-end** du projet *Application de gestion de services*.  
Elle permet d’interagir avec l’API du backend Go pour tester les principales fonctionnalités : connexion, services, réservations et administration.

---

## 🧱 Structure des fichiers

| Fichier | Rôle |
|----------|------|
| **index.html** | Page principale. Contient la structure HTML et les liens vers le CSS et le JS. |
| **css/style.css** | Fichier de styles : gère uniquement la mise en forme visuelle de la page. |
| **js/app.js** | Gère les interactions et les requêtes HTTP avec le backend (connexion, réservation, etc.). |

---

## 👤 Fonctionnement pour l’utilisateur

### 1. Connexion
- Entrer un **email** dans le champ prévu et cliquer sur **Se connecter**.  
- L’email est sauvegardé dans le navigateur (localStorage).  
- Si l’email est `admin@example.com`, les actions d’administration deviennent disponibles.

---

### 2. Voir les services
- Cliquer sur **Charger** dans la section “Services”.  
- Les services disponibles s’affichent sous forme de petites cartes grises avec leurs créneaux horaires.

---

### 3. Réserver un créneau
- Copier un **Slot ID** (identifiant d’un créneau affiché dans la liste ou créé en admin).  
- Le coller dans le champ **Slot ID** de la section “Réserver”.  
- Cliquer sur **Réserver** pour confirmer.

---

### 4. Consulter et annuler ses réservations
- Cliquer sur **Actualiser** pour afficher vos réservations.  
- Copier l’**ID de réservation** souhaité.  
- Le coller dans le champ **Reservation ID**, puis cliquer sur **Annuler**.

---

### 5. Administration (`admin@example.com`)
- **Ajouter un service** : saisir un nom, une description (optionnelle) et une durée (en minutes).  
- **Ajouter un créneau** : entrer l’ID du service, une date/heure au format `YYYY-MM-DDTHH:MM:SSZ`, et une capacité.  
- Les retours (service ou créneau créé) s’affichent sous la section “Admin”.

---

## Important
- Chaque **Service**, **Créneau (Slot)** et **Réservation** possède un **identifiant unique (ID)**.  
- Ces IDs sont affichés dans les résultats JSON ou dans les cartes de service.  
- Pour réserver, il faut copier un **Slot ID** valide.