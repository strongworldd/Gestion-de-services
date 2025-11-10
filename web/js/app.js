// --------- Helpers de session ---------
function email() { return localStorage.getItem('email') || ''; }
function setEmail(v){ localStorage.setItem('email', v || ''); updateWho(); }
function updateWho(){ document.getElementById('who').textContent = email() || 'aucun'; }

// --------- Helpers réseau ---------
async function api(path, opts={}){
  opts.headers = Object.assign({'Content-Type':'application/json'}, opts.headers||{});
  const r = await fetch(path, opts);
  let body = null;
  try { body = await r.json(); } catch(_) {}
  return { ok: r.ok, status: r.status, body };
}

async function apiGet(url, headers={}) {
  const r = await fetch(url, { headers });
  if (!r.ok) return null;
  try { return await r.json(); } catch { return null; }
}

// --------- Rendu Services (style app de base) ---------
function escapeHtml(s){
  return (s || '').replace(/[&<>"']/g, c => ({
    '&':'&amp;',
    '<':'&lt;',
    '>':'&gt;',
    '"':'&quot;',
    "'":'&#39;'
  }[c]));
}

function renderServicesFormatted(list, slotsByService) {
  const box = document.getElementById('svcList');
  if (!list || !list.length) { box.innerHTML = '<i>(aucun service)</i>'; return; }

  box.innerHTML = list.map((s, i) => {
    const slots = slotsByService[s.id] || [];
    const slotsTxt = slots.length
      ? slots.map(x => x.datetime).join(', ')
      : '(aucun)';
    // On n’a pas "type" dans le backend Go — on affiche la description si présente
    const typeOrDesc = s.description ? `(${escapeHtml(s.description)})` : '';
    return `
      <div class="svc-item">
        <div><b>#${i+1} ${escapeHtml(s.name)} ${typeOrDesc}</b></div>
        <div>Créneaux: ${escapeHtml(slotsTxt)}</div>
      </div>
    `;
  }).join('');
}

// --------- DOM refs ---------
const el = {
  loginForm:     document.getElementById('loginForm'),
  emailInput:    document.getElementById('emailInput'),

  btnLoadSvc:    document.getElementById('btnLoadServices'),
  svcList:       document.getElementById('svcList'),

  bookForm:      document.getElementById('bookForm'),
  slotIdInput:   document.getElementById('slotIdInput'),

  btnLoadMyRes:  document.getElementById('btnLoadMyRes'),
  resBox:        document.getElementById('resBox'),
  cancelForm:    document.getElementById('cancelForm'),
  resIdInput:    document.getElementById('resIdInput'),

  addSvcForm:    document.getElementById('addSvcForm'),
  svcName:       document.getElementById('svcName'),
  svcDesc:       document.getElementById('svcDesc'),
  svcDur:        document.getElementById('svcDur'),

  addSlotForm:   document.getElementById('addSlotForm'),
  slotSvcId:     document.getElementById('slotSvcId'),
  slotDt:        document.getElementById('slotDt'),
  slotCap:       document.getElementById('slotCap'),

  adminOut:      document.getElementById('adminOut'),
};

// --------- Init ---------
updateWho();

// --------- Events ---------
el.loginForm.addEventListener('submit', async (e)=>{
  e.preventDefault();
  const em = el.emailInput.value.trim();
  if(!em) return alert('Entre un email');
  await api('/auth/login', { method:'POST', body: JSON.stringify({ email: em }) });
  setEmail(em);
});

// Charger services (et slots si l’endpoint existe)
el.btnLoadSvc.addEventListener('click', async ()=>{
  const services = await apiGet('/services');
  if (!services) { el.svcList.innerHTML = '<i>(erreur chargement)</i>'; return; }

  // tente /services/:id/slots pour chaque service
  const results = await Promise.allSettled(
    services.map(s => apiGet(`/services/${s.id}/slots`))
  );
  const slotsByService = {};
  results.forEach((res, idx) => {
    const svc = services[idx];
    slotsByService[svc.id] = (res.status === 'fulfilled' && Array.isArray(res.value)) ? res.value : [];
  });

  renderServicesFormatted(services, slotsByService);
});

// Réserver
el.bookForm.addEventListener('submit', async (e)=>{
  e.preventDefault();
  const em = email(); if(!em) return alert('Connecte-toi');
  const slotId = el.slotIdInput.value.trim();
  if(!slotId) return alert('Slot ID requis');

  const {ok, body} = await api('/reservations', {
    method:'POST',
    headers:{ 'X-User-Email': em },
    body: JSON.stringify({ slotId })
  });
  if(!ok) return alert(body?.error || 'Erreur réservation');
  alert('Réservation OK: ' + body.id);
});

// Mes réservations
el.btnLoadMyRes.addEventListener('click', async ()=>{
  const em = email(); if(!em) return alert('Connecte-toi');
  const {ok, body} = await api('/reservations/me', {
    method:'GET',
    headers:{ 'X-User-Email': em }
  });
  el.resBox.textContent = ok ? JSON.stringify(body, null, 2) : '[erreur]';
});

// Annuler
el.cancelForm.addEventListener('submit', async (e)=>{
  e.preventDefault();
  const em = email(); if(!em) return alert('Connecte-toi');
  const id = el.resIdInput.value.trim(); if(!id) return alert('Reservation ID requis');
  const {ok, body} = await api('/reservations/' + id, {
    method:'DELETE',
    headers:{ 'X-User-Email': em }
  });
  if(!ok) return alert(body?.error || 'Erreur annulation');
  alert('Annulée');
});

// Admin: créer service
el.addSvcForm.addEventListener('submit', async (e)=>{
  e.preventDefault();
  const em = email();
  if(em !== 'admin@example.com') return alert('Action admin: connecte-toi en admin@example.com');

  const svc = {
    name: el.svcName.value.trim(),
    description: el.svcDesc.value.trim(),
    duration: Number(el.svcDur.value || 0) || 0
  };
  if(!svc.name) return alert('Nom requis');

  const {ok, body} = await api('/admin/services', {
    method:'POST',
    headers:{ 'X-User-Email': em },
    body: JSON.stringify(svc)
  });
  el.adminOut.textContent = ok ? ('Service créé: ' + body.id) : (body?.error || 'Erreur');
});

// Admin: ajouter créneau
el.addSlotForm.addEventListener('submit', async (e)=>{
  e.preventDefault();
  const em = email();
  if(em !== 'admin@example.com') return alert('Action admin: connecte-toi en admin@example.com');

  const sid = el.slotSvcId.value.trim();
  const dt  = el.slotDt.value.trim();
  const cap = Number(el.slotCap.value || 1) || 1;
  if(!sid || !dt) return alert('Service ID + Datetime requis');

  const {ok, body} = await api(`/admin/services/${sid}/slots`, {
    method:'POST',
    headers:{ 'X-User-Email': em },
    body: JSON.stringify({ datetime: dt, capacity: cap })
  });
  el.adminOut.textContent = ok
    ? ('Créneau ajouté. Slot ID: ' + body.id + '\nCopie cet ID pour réserver.')
    : (body?.error || 'Erreur');
});