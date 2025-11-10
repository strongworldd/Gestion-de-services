// --------- Helpers ---------
function email() { return localStorage.getItem('email') || ''; }
function setEmail(v){ localStorage.setItem('email', v || ''); updateWho(); }
function updateWho(){ document.getElementById('who').textContent = email() || 'aucun'; }

function jsonBox(id, data){
  document.getElementById(id).textContent = JSON.stringify(data, null, 2);
}

async function api(path, opts={}){
  opts.headers = Object.assign(
    {'Content-Type':'application/json'},
    opts.headers||{}
  );
  const r = await fetch(path, opts);
  let body = null;
  try { body = await r.json(); } catch(_) {}
  return { ok: r.ok, status: r.status, body };
}

// --------- DOM refs ---------
const el = {
  loginForm:     document.getElementById('loginForm'),
  emailInput:    document.getElementById('emailInput'),
  btnLoadSvc:    document.getElementById('btnLoadServices'),
  svcBox:        document.getElementById('svcBox'),

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

el.btnLoadSvc.addEventListener('click', async ()=>{
  const {ok, body} = await api('/services', { method:'GET' });
  jsonBox('svcBox', ok ? body : {error:'fail'});
});

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

el.btnLoadMyRes.addEventListener('click', async ()=>{
  const em = email(); if(!em) return alert('Connecte-toi');
  const {ok, body} = await api('/reservations/me', {
    method:'GET',
    headers:{ 'X-User-Email': em }
  });
  jsonBox('resBox', ok ? body : {error:'fail'});
});

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