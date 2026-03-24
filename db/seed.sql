INSERT OR IGNORE INTO products 
(name, price, image, cpu, ram, storage, gpu, description) 
VALUES
-- ===== Игровые ПК =====
(
  'Игровой ПК Ares',
  89990,
  '/images/pc1.png',
  'Intel Core i5-12400F',
  '16 ГБ DDR4',
  '512 ГБ SSD',
  'NVIDIA GeForce RTX 3060',
  'Отличный игровой ПК для Full HD и работы.'
),
(
  'Игровой ПК Nitro X',
  129990,
  '/images/pc2.png',
  'Intel Core i7-13700F',
  '32 ГБ DDR5',
  '1 ТБ SSD',
  'NVIDIA GeForce RTX 4070',
  'Мощная сборка для современных игр и стриминга.'
),
(
  'Игровой ПК Shadow',
  109990,
  '/images/pc3.png',
  'AMD Ryzen 7 5800X',
  '16 ГБ DDR4',
  '1 ТБ SSD',
  'NVIDIA GeForce RTX 3070',
  'Баланс цены и производительности для современных игр.'
),

-- ===== Офисные ПК =====
(
  'Офисный ПК Office Pro',
  39990,
  '/images/pc4.png',
  'Intel Core i5-11400',
  '16 ГБ DDR4',
  '512 ГБ SSD',
  'Встроенная графика Intel UHD',
  'Надёжный ПК для работы, браузера и офисных задач.'
),
(
  'Офисный ПК Compact',
  29990,
  '/images/pc5.png',
  'Intel Core i3-10100',
  '8 ГБ DDR4',
  '256 ГБ SSD',
  'Встроенная графика Intel UHD',
  'Тихий и компактный офисный компьютер для дома и работы.'
),
(
  'Офисный ПК Mini Box',
  25990,
  '/images/pc6.png',
  'AMD Ryzen 3 4100',
  '8 ГБ DDR4',
  '256 ГБ SSD',
  'Встроенная графика',
  'Мини-ПК, который легко помещается на столе или за монитором.'
);