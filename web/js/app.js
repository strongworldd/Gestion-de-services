// --------- Helpers de session ---------

function getEmail() {
  return localStorage.getItem('email') || '';
}

function email() { return getEmail(); }

function setEmail(value) {
  localStorage.setItem('email', value || '');
  updateWho();
}

function updateWho() {
  const whoElement = document.getElementById('who');
  if (!whoElement) return;
  whoElement.textContent = getEmail() || 'aucun';
}

function logout() {
  // Supprime l'email stocké
  localStorage.removeItem('email');

  // Met à jour le texte "Connecté: ..."
  updateWho();

  // Vide le champ email
  const emailInput = document.getElementById('emailInput');
  if (emailInput) {
    emailInput.value = '';
  }

  // Recharge la page pour repartir sans utilisateur connecté
  window.location.reload();
}
// --------- Helpers réseau ---------

async function api(path, opts = {}) {
  opts.headers = Object.assign(
    { 'Content-Type': 'application/json' },
    opts.headers || {}
  );

  const response = await fetch(path, opts);
  let body = null;

  try {
    body = await response.json();
  } catch {
    // Ignorer l'erreur si la réponse n'est pas en JSON
  }

  return {
    ok: response.ok,
    status: response.status,
    body,
  };
}

async function apiGet(url, headers = {}) {
  const response = await fetch(url, { headers });

  if (!response.ok) {
    return null;
  }

  try {
    return await response.json();
  } catch {
    return null;
  }
}

// --------- Helper : Échappement HTML ---------

function escapeHtml(str) {
  return (str || '').replace(/[&<>"']/g, (char) => {
    const map = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#39;',
    };
    return map[char];
  });
}

// --------- Rendu des services ---------

function renderServicesFormatted(services, slotsByService) {
  const container = document.getElementById('svcList');

  if (!services || services.length === 0) {
    container.innerHTML = '<i>(aucun service)</i>';
    return;
  }

  container.innerHTML = services
    .map((service, index) => {
      const slots = slotsByService[service.id] || [];
      const slotsText = slots.length
        ? slots.map((slot) => slot.datetime).join(', ')
        : '(aucun)';

      const descriptionText = service.description
        ? `(${escapeHtml(service.description)})`
        : '';

      return `
        <div class="svc-item">
          <div>
            <b>#${index + 1} ${escapeHtml(service.name)} ${descriptionText}</b>
          </div>
          <div>Créneaux : ${escapeHtml(slotsText)}</div>
        </div>
      `;
    })
    .join('');
}

// --------- Références DOM ---------
const el = {
  // Connexion utilisateur
  loginForm: document.getElementById('loginForm'),
  emailInput: document.getElementById('emailInput'),
  logoutBtn: document.getElementById('logoutBtn'),

  // Chargement et affichage des services
  btnLoadSvc: document.getElementById('btnLoadServices'),
  svcList: document.getElementById('svcList'),

  // Réservation
  bookForm: document.getElementById('bookForm'),
  slotIdInput: document.getElementById('slotIdInput'),

  // Consultation / annulation de réservation
  btnLoadMyRes: document.getElementById('btnLoadMyRes'),
  resBox: document.getElementById('resBox'),
  cancelForm: document.getElementById('cancelForm'),
  resIdInput: document.getElementById('resIdInput'),

  // Gestion des services (admin)
  addSvcForm: document.getElementById('addSvcForm'),
  svcName: document.getElementById('svcName'),
  svcDesc: document.getElementById('svcDesc'),
  svcDur: document.getElementById('svcDur'),

  // Gestion des créneaux (admin)
  addSlotForm: document.getElementById('addSlotForm'),
  slotSvcId: document.getElementById('slotSvcId'),
  slotDt: document.getElementById('slotDt'),
  slotCap: document.getElementById('slotCap'),

  // Zone d’affichage admin
  adminOut: document.getElementById('adminOut'),
};

// --------- Init ---------
updateWho();

// --------- Events ---------
el.loginForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  const userEmail = el.emailInput.value.trim();
  if (!userEmail) return alert('Entre un email');

  await api('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email: userEmail }),
  });

  setEmail(userEmail);
});

// Déconnexion
if (el.logoutBtn) {
  el.logoutBtn.addEventListener('click', () => {
    logout();
  });
}

// --------- Charger les services (et slots si l’endpoint existe) ---------
el.btnLoadSvc.addEventListener('click', async () => {
  const services = await apiGet('/services');

  if (!services) {
    el.svcList.innerHTML = '<i>(erreur chargement)</i>';
    return;
  }

  // Tente de charger /services/:id/slots pour chaque service
  const results = await Promise.allSettled(
    services.map((service) => apiGet(`/services/${service.id}/slots`))
  );

  const slotsByService = {};

  results.forEach((result, index) => {
    const service = services[index];
    const isOk = result.status === 'fulfilled' && Array.isArray(result.value);
    slotsByService[service.id] = isOk ? result.value : [];
  });

  renderServicesFormatted(services, slotsByService);
});

// --------- Réserver ---------
el.bookForm.addEventListener('submit', async (e) => {
  e.preventDefault();

  const userEmail = email();
  if (!userEmail) {
    alert('Connecte-toi');
    return;
  }

  const slotId = el.slotIdInput.value.trim();
  if (!slotId) {
    alert('Slot ID requis');
    return;
  }

  const { ok, body } = await api('/reservations', {
    method: 'POST',
    headers: { 'X-User-Email': userEmail },
    body: JSON.stringify({ slotId }),
  });

  if (!ok) {
    alert(body?.error || 'Erreur réservation');
    return;
  }

  alert(`Réservation OK : ${body.id}`);
});

// --------- Mes réservations ---------
el.btnLoadMyRes.addEventListener('click', async () => {
  const userEmail = email();

  if (!userEmail) {
    alert('Connecte-toi');
    return;
  }

  const { ok, body } = await api('/reservations/me', {
    method: 'GET',
    headers: { 'X-User-Email': userEmail },
  });

  el.resBox.textContent = ok
    ? JSON.stringify(body, null, 2)
    : '[erreur]';
});

// --------- Annuler ---------
el.cancelForm.addEventListener('submit', async (e) => {
  e.preventDefault();

  const userEmail = email();
  if (!userEmail) {
    alert('Connecte-toi');
    return;
  }

  const reservationId = el.resIdInput.value.trim();
  if (!reservationId) {
    alert('Reservation ID requis');
    return;
  }

  const { ok, body } = await api(`/reservations/${reservationId}`, {
    method: 'DELETE',
    headers: { 'X-User-Email': userEmail },
  });

  if (!ok) {
    alert(body?.error || 'Erreur annulation');
    return;
  }

  alert('Annulée');
});

// --------- Admin : créer un service ---------
el.addSvcForm.addEventListener('submit', async (e) => {
  e.preventDefault();

  const userEmail = email();
  if (userEmail !== 'admin@example.com') {
    alert('Action admin : connecte-toi en admin@example.com');
    return;
  }

  const service = {
    name: el.svcName.value.trim(),
    description: el.svcDesc.value.trim(),
    duration: Number(el.svcDur.value || 0) || 0,
  };

  if (!service.name) {
    alert('Nom requis');
    return;
  }

  const { ok, body } = await api('/admin/services', {
    method: 'POST',
    headers: { 'X-User-Email': userEmail },
    body: JSON.stringify(service),
  });

  el.adminOut.textContent = ok
    ? `Service créé : ${body.id}`
    : body?.error || 'Erreur';
});

// --------- Admin : ajouter un créneau ---------
el.addSlotForm.addEventListener('submit', async (e) => {
  e.preventDefault();

  const userEmail = email();
  if (userEmail !== 'admin@example.com') {
    alert('Action admin : connecte-toi en admin@example.com');
    return;
  }

  const serviceId = el.slotSvcId.value.trim();
  const dateTime = el.slotDt.value.trim();
  const capacity = Number(el.slotCap.value || 1) || 1;

  if (!serviceId || !dateTime) {
    alert('Service ID et Date/Heure requis');
    return;
  }

  const { ok, body } = await api(`/admin/services/${serviceId}/slots`, {
    method: 'POST',
    headers: { 'X-User-Email': userEmail },
    body: JSON.stringify({ datetime: dateTime, capacity }),
  });

  el.adminOut.textContent = ok
    ? `Créneau ajouté. Slot ID : ${body.id}\nCopie cet ID pour réserver.`
    : body?.error || 'Erreur';
});