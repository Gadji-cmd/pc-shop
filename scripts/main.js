// Запускаемся после загрузки DOM
document.addEventListener('DOMContentLoaded', () => {
  const path = window.location.pathname;

  // Инициализация системы авторизации
  initAuthSystem();

  // Главная — популярные ПК
  if (path === '/') {
    initHomePage();
  }

  // Каталог — игровые / офисные
  if (path === '/products') {
    initProductsPage();
  }

  // Контакты — инициализация карты
  if (path === '/contacts') {
    initContactsPage();
  }

  // Профиль
  if (path === '/profile') {
    initProfilePage();
  }

  // Корзина
  if (path === '/cart') {
    initCartPage();
  }
});

//  СИСТЕМА АВТОРИЗАЦИИ 

let currentUser = null;

async function initAuthSystem() {
  const authDialog = document.getElementById('authDialog');
  const regDialog = document.getElementById('regDialog');
  const btnOpenAuth = document.getElementById('btnOpenAuth');
  const btnToRegister = document.getElementById('btnToRegister');
  const btnCancelLogin = document.getElementById('btnCancelLogin');
  const btnCancelRegister = document.getElementById('btnCancelRegister');
  const formLogin = document.getElementById('formLogin');
  const formRegister = document.getElementById('formRegister');

  // Проверяем авторизацию при загрузке
  await checkAuth();

  if (!authDialog || !regDialog) return;

  // Открытие диалога входа
  if (btnOpenAuth) {
    btnOpenAuth.addEventListener('click', (e) => {
      e.preventDefault();
      if (currentUser) {
        window.location.href = '/profile';
      } else {
        authDialog.showModal();
      }
    });
  }

  // Переход к регистрации
  if (btnToRegister) {
    btnToRegister.addEventListener('click', () => {
      authDialog.close();
      regDialog.showModal();
    });
  }

  // Отмена входа
  if (btnCancelLogin) {
    btnCancelLogin.addEventListener('click', () => {
      authDialog.close();
    });
  }

  // Отмена регистрации
  if (btnCancelRegister) {
    btnCancelRegister.addEventListener('click', () => {
      regDialog.close();
    });
  }

  // Обработка входа
  if (formLogin) {
    formLogin.addEventListener('submit', async (e) => {
      e.preventDefault();
      const formData = new FormData(formLogin);
      const email = formData.get('email');
      const password = formData.get('password');

      try {
        const response = await fetch('/api/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ email, password })
        });

        if (response.ok) {
          const user = await response.json();
          currentUser = user;
          showAuthMessage('authMsg', 'Вход выполнен успешно!', 'success');
          updateAuthButton();
          setTimeout(() => {
            authDialog.close();
            formLogin.reset();
          }, 1000);
        } else {
          showAuthMessage('authMsg', 'Неверный email или пароль', 'error');
        }
      } catch (error) {
        showAuthMessage('authMsg', 'Ошибка сети', 'error');
      }
    });
  }

  // Обработка регистрации
  if (formRegister) {
    formRegister.addEventListener('submit', async (e) => {
      e.preventDefault();
      const formData = new FormData(formRegister);
      const email = formData.get('email');
      const password = formData.get('password');
      const name = email.split('@')[0]; // Простое имя из email

      try {
        const response = await fetch('/api/register', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ email, password, name })
        });

        if (response.ok) {
          const user = await response.json();
          currentUser = user;
          showAuthMessage('regMsg', 'Регистрация успешна!', 'success');
          updateAuthButton();
          setTimeout(() => {
            regDialog.close();
            formRegister.reset();
            authDialog.showModal();
          }, 1000);
        } else {
          const error = await response.text();
          showAuthMessage('regMsg', 'Ошибка регистрации: ' + error, 'error');
        }
      } catch (error) {
        showAuthMessage('regMsg', 'Ошибка сети', 'error');
      }
    });
  }
}

async function checkAuth() {
  try {
    const response = await fetch('/api/user');
    if (response.ok) {
      currentUser = await response.json();
      updateAuthButton();
      return true;
    }
  } catch (error) {
    console.error('Ошибка проверки авторизации:', error);
  }
  currentUser = null;
  updateAuthButton();
  return false;
}

function updateAuthButton() {
  const btnOpenAuth = document.getElementById('btnOpenAuth');
  if (btnOpenAuth) {
    if (currentUser) {
      btnOpenAuth.textContent = 'Профиль';
      btnOpenAuth.style.backgroundColor = '#10b981';
    } else {
      btnOpenAuth.textContent = 'Вход';
      btnOpenAuth.style.backgroundColor = '';
    }
  }
}

function showAuthMessage(elementId, message, type) {
  const element = document.getElementById(elementId);
  if (element) {
    element.textContent = message;
    element.style.color = type === 'success' ? '#10b981' : '#ef4444';
  }
}

//  КАРТА 

function initContactsPage() {
  if (typeof ymaps !== 'undefined') {
    ymaps.ready(() => {
      const map = new ymaps.Map('map', {
        center: [59.925101, 30.318007],
        zoom: 12
      });

      const placemark = new ymaps.Placemark([59.879146, 30.275893], {
        hintContent: 'PC Shop',
        balloonContent: 'г. Санкт-Петербург, ул. Зайцева, д. 39'
      });

      map.geoObjects.add(placemark);
    });
  } else {
    console.warn('Yandex Maps API не загружен');
  }
}

// ПРОФИЛЬ 

async function initProfilePage() {
  if (!await checkAuth()) {
    window.location.href = '/';
    return;
  }

  const profileInfo = document.getElementById('profileInfo');
  if (profileInfo) {
    profileInfo.innerHTML = `
      <div class="profile-card">
        <h3>${currentUser.name}</h3>
        <p><strong>Email:</strong> ${currentUser.email}</p>
        <p><strong>Зарегистрирован:</strong> ${new Date(currentUser.created_at).toLocaleDateString('ru-RU')}</p>
        <button id="btnLogout" class="btn primary">Выйти</button>
      </div>
    `;

    document.getElementById('btnLogout').addEventListener('click', async () => {
      await fetch('/api/logout', { method: 'POST' });
      currentUser = null;
      updateAuthButton();
      window.location.href = '/';
    });
  }
}

// КОРЗИНА 

async function initCartPage() {
  if (!await checkAuth()) {
    window.location.href = '/';
    return;
  }

  await loadCart();
}

async function loadCart() {
  try {
    const response = await fetch('/api/cart');
    if (!response.ok) {
      throw new Error('Ошибка загрузки корзины');
    }

    const cartItems = await response.json();
    const container = document.getElementById('cartItems');
    
    if (!container) return;

    if (cartItems.length === 0) {
      container.innerHTML = '<p class="empty-cart">Корзина пуста</p>';
      return;
    }

    let total = 0;
    container.innerHTML = cartItems.map(item => {
      const itemTotal = item.product.price * item.quantity;
      total += itemTotal;
      
      return `
        <div class="cart-item">
          <img src="${item.product.image}" alt="${item.product.name}">
          <div class="cart-item-info">
            <h4>${item.product.name}</h4>
            <p class="price">${formatPrice(item.product.price)} ₽ × ${item.quantity}</p>
            <p class="item-total">${formatPrice(itemTotal)} ₽</p>
          </div>
          <button class="btn-remove" data-item-id="${item.id}">✕</button>
        </div>
      `;
    }).join('') + `
      <div class="cart-total">
        <h3>Итого: ${formatPrice(total)} ₽</h3>
        <button class="btn primary" id="btnCheckout">Оформить заказ</button>
      </div>
    `;

    // Обработчики удаления
    document.querySelectorAll('.btn-remove').forEach(btn => {
      btn.addEventListener('click', async (e) => {
        const itemId = e.target.getAttribute('data-item-id');
        await removeFromCart(itemId);
        await loadCart();
      });
    });

    document.getElementById('btnCheckout')?.addEventListener('click', () => {
      alert('Заказ оформлен! Спасибо за покупку!');
      // Здесь можно добавить логику оформления заказа
    });

  } catch (error) {
    console.error('Ошибка загрузки корзины:', error);
    document.getElementById('cartItems').innerHTML = '<p class="error">Ошибка загрузки корзины</p>';
  }
}

async function removeFromCart(itemId) {
  try {
    const response = await fetch(`/api/cart/remove/${itemId}`, {
      method: 'DELETE'
    });
    
    if (!response.ok) {
      throw new Error('Ошибка удаления из корзины');
    }
  } catch (error) {
    console.error('Ошибка удаления из корзины:', error);
    alert('Ошибка при удалении товара из корзины');
  }
}

//  ОБЩИЙ ЗАПРОС ТОВАРОВ 

async function fetchProducts() {
  try {
    const response = await fetch('/api/products');
    if (!response.ok) {
      throw new Error('Ошибка ответа сервера: ' + response.status);
    }
    const products = await response.json();
    if (!Array.isArray(products)) {
      throw new Error('Некорректный формат данных от сервера');
    }
    return products;
  } catch (error) {
    console.error('Ошибка загрузки товаров:', error);
    return [];
  }
}

//  ГЛАВНАЯ 

async function initHomePage() {
  const container = document.getElementById('products');
  if (!container) return;

  const products = await fetchProducts();
  if (products.length === 0) {
    container.textContent = 'Популярные товары временно недоступны.';
    return;
  }

  const popular = products.slice(0, 3);
  popular.forEach((product) => {
    const card = createProductCard(product);
    container.appendChild(card);
  });
}

// КАТАЛОГ 

async function initProductsPage() {
  const gamingEl = document.getElementById('catalogGaming');
  const officeEl = document.getElementById('catalogOffice');
  const statusEl = document.getElementById('catalogStatus');

  if (!gamingEl || !officeEl || !statusEl) {
    return;
  }

  statusEl.textContent = 'Загрузка товаров...';
  const products = await fetchProducts();

  if (products.length === 0) {
    statusEl.textContent = 'Пока нет товаров 😔';
    return;
  }

  const gaming = [];
  const office = [];

  products.forEach((product) => {
    if (isGamingProduct(product)) {
      gaming.push(product);
    } else {
      office.push(product);
    }
  });

  statusEl.textContent = '';

  // Рендерим игровые
  if (gaming.length > 0) {
    gaming.forEach((p) => {
      const card = createProductCard(p);
      gamingEl.appendChild(card);
    });
  } else {
    gamingEl.textContent = 'Игровых ПК пока нет.';
  }

  // Рендерим офисные
  if (office.length > 0) {
    office.forEach((p) => {
      const card = createProductCard(p);
      officeEl.appendChild(card);
    });
  } else {
    officeEl.textContent = 'Офисных ПК пока нет.';
  }
}

function isGamingProduct(product) {
  const name = (product.name || '').toLowerCase();
  return name.includes('игров');
}

// СОЗДАНИЕ КАРТОЧКИ 

function createProductCard(product) {
  const card = document.createElement('article');
  card.className = 'product-card';

  const name = product.name || 'Без названия';
  const price = typeof product.price === 'number' ? product.price : 0;
  const image = product.image || '/images/pc1.png';
  const cpu = product.cpu || '';
  const ram = product.ram || '';
  const storage = product.storage || '';
  const gpu = product.gpu || '';
  const description = product.description || '';

  card.innerHTML = `
    <img src="${image}" alt="${name}">
    <h3>${name}</h3>
    <p class="price">${formatPrice(price)} ₽</p>
    <ul class="specs">
      ${cpu ? `<li><span>CPU:</span> ${cpu}</li>` : ''}
      ${ram ? `<li><span>RAM:</span> ${ram}</li>` : ''}
      ${storage ? `<li><span>Диск:</span> ${storage}</li>` : ''}
      ${gpu ? `<li><span>GPU:</span> ${gpu}</li>` : ''}
    </ul>
    ${description ? `<p class="product-description">${description}</p>` : ''}
    <button class="btn-buy" data-product-id="${product.id}">Купить</button>
  `;

  // Обработчик кнопки "Купить"
  const buyBtn = card.querySelector('.btn-buy');
  buyBtn.addEventListener('click', async () => {
    if (!await checkAuth()) {
      alert('Для добавления товаров в корзину необходимо авторизоваться');
      return;
    }

    try {
      const response = await fetch('/api/cart/add', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          product_id: product.id,
          quantity: 1
        })
      });

      if (response.ok) {
        alert('Товар добавлен в корзину!');
      } else {
        alert('Ошибка при добавлении товара в корзину');
      }
    } catch (error) {
      console.error('Ошибка:', error);
      alert('Ошибка сети');
    }
  });

  return card;
}

function formatPrice(value) {
  if (typeof value !== 'number') {
    return value;
  }
  return value.toLocaleString('ru-RU');
}