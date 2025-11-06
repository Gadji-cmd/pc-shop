CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    specs TEXT NOT NULL,
    price INTEGER NOT NULL,
    image TEXT
);

INSERT INTO products (title, specs, price, image) VALUES
('Игровой ПК Nitro X', 'Ryzen 5 5600, 16GB RAM, RTX 3060, SSD 512GB', 89990, '/public/img/pc1.jpg'),
('Игровой ПК Nitro 2X', 'Intel i7, 32GB RAM, RTX 3060, SSD 1024GB', 134990, '/public/img/pc2.jpg'),
('Ультра ПК Creator', 'Ryzen 7 7800X, 32GB RAM, RTX 4070, SSD 1TB', 169990, '/public/img/pc3.jpg'),
('Компактный ПК Mini', 'Intel N100, 8GB RAM, SSD 256GB', 25990, '/public/img/pc4.jpg');
