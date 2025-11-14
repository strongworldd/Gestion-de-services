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

// --------- Cache services/slots pour réutilisation ---------
const svcCache = {
  services: [],
  slotsByService: {},
  slotCatalog: {},
  slotCatalogLoaded: false,
};

function updateSlotCatalogCache(services = [], slotsByService = {}) {
  svcCache.services = services;
  svcCache.slotsByService = slotsByService;
  svcCache.slotCatalog = {};
  svcCache.slotCatalogLoaded = true;

  services.forEach((service) => {
    const serviceLabel = service.description
      ? `${service.name} (${service.description})`
      : service.name;
    const slots = slotsByService[service.id] || [];
    slots.forEach((slot) => {
      if (!slot || !slot.id) return;
      svcCache.slotCatalog[slot.id] = {
        serviceLabel,
        datetime: slot.datetime || '',
      };
    });
  });
}

async function fetchServicesWithSlots() {
  const services = await apiGet('/services');
  if (!Array.isArray(services)) {
    return null;
  }

  const results = await Promise.allSettled(
    services.map((service) => apiGet(`/services/${service.id}/slots`))
  );

  const slotsByService = {};

  results.forEach((result, index) => {
    const service = services[index];
    const isOk = result.status === 'fulfilled' && Array.isArray(result.value);
    slotsByService[service.id] = isOk ? result.value : [];
  });

  return { services, slotsByService };
}

async function ensureSlotCatalog() {
  if (svcCache.slotCatalogLoaded) {
    return svcCache.slotCatalog;
  }

  const svcData = await fetchServicesWithSlots();
  if (!svcData) {
    return {};
  }

  updateSlotCatalogCache(svcData.services, svcData.slotsByService);
  return svcCache.slotCatalog;
}

function formatReservation(reservation, slotCatalog = {}) {
  const slotInfo = slotCatalog[reservation.slotId] || null;
  const serviceLabel = slotInfo
    ? slotInfo.serviceLabel
    : `Créneau ${reservation.slotId || 'inconnu'}`;
  const slotDatetime = slotInfo && slotInfo.datetime
    ? slotInfo.datetime
    : 'Date inconnue';
  const createdAt = reservation.createdAt
    ? `Réservé le ${reservation.createdAt}`
    : '';

  const createdHtml = createdAt
    ? `<div class="muted">${escapeHtml(createdAt)}</div>`
    : '';

  return `
    <div class="res-item">
      <div><b>${escapeHtml(serviceLabel)}</b></div>
      <div>Créneau : ${escapeHtml(slotDatetime)}</div>
      <div class="muted">ID réservation : ${escapeHtml(reservation.id || '')}</div>
      ${createdHtml}
    </div>
  `;
}

function renderReservations(reservations, slotCatalog) {
  if (!el.resBox) return;

  if (!Array.isArray(reservations) || reservations.length === 0) {
    el.resBox.innerHTML = '<i>(aucune réservation)</i>';
    return;
  }

  const html = reservations
    .map((res) => formatReservation(res, slotCatalog))
    .join('')
    .trim();

  el.resBox.innerHTML = html || '<i>(créneaux indisponibles)</i>';
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
        ? slots
            .map((slot) => {
              const slotId = slot.id
                ? escapeHtml(slot.id)
                : '[id inconnu]';
              const slotDateTime = slot.datetime
                ? escapeHtml(slot.datetime)
                : '[date inconnue]';
              return `${slotId} – ${slotDateTime}`;
            })
            .join(', ')
        : '(aucun)';

      const descriptionText = service.description
        ? `(${escapeHtml(service.description)})`
        : '';

      return `
        <div class="svc-item">
          <div>
            <b>#${index + 1} ${escapeHtml(service.name)} ${descriptionText}</b>
          </div>
          <div>Créneaux : ${slotsText}</div>
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
  const svcData = await fetchServicesWithSlots();

  if (!svcData) {
    el.svcList.innerHTML = '<i>(erreur chargement)</i>';
    return;
  }

  updateSlotCatalogCache(svcData.services, svcData.slotsByService);
  renderServicesFormatted(svcData.services, svcData.slotsByService);
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

  if (!ok) {
    el.resBox.innerHTML = '<i>(erreur)</i>';
    return;
  }

  if (!Array.isArray(body) || body.length === 0) {
    el.resBox.innerHTML = '<i>(aucune réservation)</i>';
    return;
  }

  const slotCatalog = await ensureSlotCatalog();
  renderReservations(body, slotCatalog);
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
