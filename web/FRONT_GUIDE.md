# Mini Booking – Guide d’utilisation du front

## 🎯 Objectif
Ce document explique le fonctionnement de la partie **front-end** (fichiers `index.html` et `app.js`) du projet Mini Booking.  
Ce front permet de tester les principales fonctionnalités utilisateur et admin via une interface web simple.

---

## 🧱 Structure des fichiers
| Fichier | Rôle |
|----------|------|
| **index.html** | Interface web principale (HTML + un peu de style). Contient les formulaires pour se connecter, réserver, consulter et gérer les services. |
| **app.js** | Code JavaScript qui gère toutes les interactions et envoie les requêtes HTTP vers le backend Go (`fetch()`). |

---

## 👤 Fonctionnement pour l’utilisateur

### 1. Connexion
- Saisir un **email** dans le champ prévu.  
- Cliquer sur **Se connecter**.  
- L’email est sauvegardé dans le navigateur (localStorage).  
- Si l’email est `admin@example.com`, les actions d’administration deviennent disponibles.

---

### 2. Voir les services
- Cliquer sur le bouton **Charger** dans la section “Services”.  
- Les services disponibles s’affichent au format JSON.

---

### 3. Réserver un créneau
- Récupérer un **Slot ID** (identifiant du créneau, visible après création côté admin).  
- Saisir cet identifiant dans le champ “Slot ID”.  
- Cliquer sur **Réserver** → une confirmation s’affiche si la réservation est acceptée.

---

### 4. Consulter et annuler ses réservations
- Cliquer sur **Actualiser** pour voir la liste de vos réservations.  
- Copier l’**ID de réservation** voulu.  
- Coller cet ID dans le champ “Reservation ID” puis cliquer sur **Annuler**.

---

### 5. Partie administration (réservée à `admin@example.com`)
- **Créer un service** : renseigner un nom, une description, et une durée (facultatif).  
- **Ajouter un créneau** : indiquer le Service ID, une date/heure au format RFC3339 (`2025-12-20T14:00:00Z`) et la capacité.  
- L’API renvoie les informations du service ou du créneau créé, dont les identifiants utiles (Service ID ou Slot ID).

---

## 🧩 À retenir
- Chaque élément (service, créneau, réservation) possède un **ID unique** généré par le backend.  
- Ces IDs apparaissent dans les retours JSON affichés sur la page.  
- Pour réserver, il faut **copier le Slot ID** affiché après l’ajout d’un créneau.  

---

📌 *Ce front minimaliste sert uniquement à tester le bon fonctionnement de l’API côté Go.  
Il ne contient aucune logique serveur — tout passe par les requêtes HTTP du backend.*