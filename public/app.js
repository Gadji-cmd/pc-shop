const api = {
  products: "/api/products",
  product: (id) => `/api/products/${id}`,
  login: "/api/login",
  register: "/api/register",
  order: "/api/order",
};

/* ---------- Главная: карточки ---------- */
async function loadProducts() {
  const cont = document.querySelector('#products');
  if (!cont) return;
  const res = await fetch(api.products);
  const list = await res.json();
  cont.innerHTML = list.map(renderCard).join('');
}
function renderCard(p){
  return `
    <div class="card">
      <img src="${p.image}" alt="${p.title}">
      <h3>${p.title}</h3>
      <p>${p.specs}</p>
      <p><strong>${p.price}</strong> ₽</p>
      <a class="btn" href="/public/product.html?id=${p.id}">Подробнее</a>
    </div>
  `;
}

/* ---------- Страница /public/product.html ----------
   Если есть id -> показываем карточку товара.
   Если id нет -> показываем каталог всех товаров + заголовок "Каталог".
--------------------------------------------------- */
async function loadCatalogOrProductPage(){
  const params = new URLSearchParams(location.search);
  const id = params.get('id');

  const catalogEl = document.querySelector('#catalog');
  const productCard = document.querySelector('#productCard');
  const pageTitle = document.querySelector('#pageTitle');

  if (!catalogEl && !productCard) return; // не эта страница

  if (!id) {
    // Режим КАТАЛОГА
    pageTitle.textContent = 'Каталог';
    productCard.style.display = 'none';
    catalogEl.style.display = 'grid';

    const res = await fetch(api.products);
    const list = await res.json();
    catalogEl.innerHTML = list.map(renderCard).join('');
    return;
  }

  // Режим КАРТОЧКИ ТОВАРА
  pageTitle.textContent = 'Товар';
  catalogEl.style.display = 'none';
  productCard.style.display = 'block';

  const h = (sel) => document.querySelector(sel);
  const res = await fetch(api.product(id));
  if (!res.ok) { h('#pTitle').textContent = 'Товар не найден'; return; }
  const p = await res.json();

  h('#pTitle').textContent = p.title;
  h('#pSpecs').textContent = p.specs;
  h('#pPrice').textContent = p.price;
  h('#pImage').src = p.image;

  // Покупка: показываем форму только по нажатию «Купить»
  const btnShowBuy = document.querySelector('#btnShowBuy');
  const buyBlock   = document.querySelector('#buyBlock');
  const formBuy    = document.querySelector('#formBuy');
  const buyMsg     = document.querySelector('#buyMsg');

  const isAuthed = document.cookie.split('; ').some(c => c.startsWith('session='));

  btnShowBuy?.addEventListener('click', () => {
    if (!isAuthed) {
      buyMsg.textContent = 'Чтобы оформить заказ — войдите в аккаунт (кнопка «Войти» в шапке).';
      buyMsg.style.color = '#ef4444';
      const dlg = document.querySelector('#authDialog');
      if (dlg && typeof dlg.showModal === 'function') dlg.showModal();
      return;
    }
    btnShowBuy.style.display = 'none';
    buyBlock.style.display = 'block';
    buyMsg.textContent = '';
  });

  formBuy?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const qty = new FormData(formBuy).get('qty') || 1;
    const r = await fetch(api.order, { method: 'POST' });
    if (r.ok) { buyMsg.textContent = `Заказ принят: ${qty} шт.`; buyMsg.style.color = '#22c55e'; }
    else { buyMsg.textContent = 'Нужно войти в аккаунт'; buyMsg.style.color = '#ef4444'; }
  });
}

/* ---------- Авторизация/Регистрация ---------- */
function initAuth() {
  const dlgAuth = document.querySelector('#authDialog');
  const dlgReg  = document.querySelector('#regDialog');

  document.querySelector('#btnOpenAuth')?.addEventListener('click', () => dlgAuth.showModal());
  document.querySelector('#btnToRegister')?.addEventListener('click', () => { dlgAuth.close(); dlgReg.showModal(); });
  document.querySelector('#btnCancelLogin')?.addEventListener('click', () => dlgAuth.close());
  document.querySelector('#btnCancelRegister')?.addEventListener('click', () => dlgReg.close());

  const formLogin = document.querySelector('#formLogin');
  formLogin?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const data = Object.fromEntries(new FormData(formLogin));
    const r = await fetch(api.login, { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify(data) });
    const msg = document.querySelector('#authMsg');
    if (r.ok) { msg.textContent = 'Успешный вход!'; msg.style.color = '#22c55e'; setTimeout(() => dlgAuth.close(), 700); }
    else { msg.textContent = 'Неверные данные'; msg.style.color = '#ef4444'; }
  });

  const formRegister = document.querySelector('#formRegister');
  formRegister?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const data = Object.fromEntries(new FormData(formRegister));
    const r = await fetch(api.register, { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify(data) });
    const msg = document.querySelector('#regMsg');
    if (r.ok) { msg.textContent = 'Аккаунт создан, теперь войдите'; msg.style.color = '#22c55e'; }
    else { msg.textContent = 'Email уже существует'; msg.style.color = '#ef4444'; }
  });
}

/* ---------- Карта (контакты) ---------- */
function initMap() {
  const el = document.getElementById('map');
  if (!el || !window.L) return;
  const center = [59.879146, 30.275893];
  const map = L.map('map').setView(center, 12);
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', { attribution: '&copy; OpenStreetMap' }).addTo(map);
  L.marker(center).addTo(map).bindPopup('PC Shop — ждём вас!');
}

/* ---------- Небольшая анимация карточек ---------- */
document.addEventListener('mouseover', (e) => {
  const card = e.target.closest('.card');
  if (card) { card.style.transform = 'translateY(-2px)'; card.style.transition = 'transform .15s'; }
});
document.addEventListener('mouseout', (e) => {
  const card = e.target.closest('.card');
  if (card) { card.style.transform = ''; }
});

/* ---------- Init ---------- */
window.addEventListener('DOMContentLoaded', () => {
  loadProducts();              // главная (если есть #products)
  loadCatalogOrProductPage();  // /public/product.html — каталог или карточка
  initAuth();
  initMap();
});
