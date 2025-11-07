package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// --- статические файлы (вшиваем папку public) ---
//go:embed public
var publicFS embed.FS

var db *sql.DB

func main() {
	// --- БД: путь и подключение ---
	dsn := defaultDSN() // file:/data/pcshop.db?... если есть /data (Render Disk), иначе локально
	justCreated := ensureDBFileCreated(dsn) // true, если файла не было и мы его создаём впервые

	var err error
	db, err = sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Применим schema.sql только при первом создании БД
	if justCreated {
		if _, err := os.Stat("schema.sql"); err == nil {
			if b, rerr := os.ReadFile("schema.sql"); rerr == nil {
				if _, e := db.Exec(string(b)); e != nil {
					log.Println("schema apply warning:", e)
				}
			}
		}
	}

	// Сид стартовых товаров, если таблица пустая (не создаёт дублей)
	seedProductsIfEmpty()

	// --- Маршруты API ---
	mux := http.NewServeMux()
	mux.HandleFunc("/api/products", handleProducts)
	mux.HandleFunc("/api/products/", handleProductByID)
	mux.HandleFunc("/api/register", handleRegister)
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/order", handleOrder)

	// Статические файлы
	mux.Handle("/", http.FileServer(http.FS(publicFS)))

	// Порт: локально 8080, на хостингах (Render) берём $PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           logging(cors(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("PC Shop running on port", port)
	log.Fatal(srv.ListenAndServe())
}

// Выбираем DSN: если есть /data (диск Render/Railway) — пишем туда; иначе локальный файл.
func defaultDSN() string {
	if _, err := os.Stat("/data"); err == nil {
		return "file:/data/pcshop.db?cache=shared&_pragma=busy_timeout=10000"
	}
	return "file:pcshop.db?cache=shared&_pragma=busy_timeout=10000"
}

// Определяем, создан ли файл БД впервые. Если не существовал — создаём пустой и вернём true.
func ensureDBFileCreated(dsn string) bool {
	// DSN вида "file:/path/to/db.sqlite?params"
	path := strings.TrimPrefix(dsn, "file:")
	if i := strings.IndexByte(path, '?'); i >= 0 {
		path = path[:i]
	}
	// относительный путь превращаем в чистый
	path = filepath.Clean(path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// создаём папку, если нужно
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		// создаём пустой файл
		f, cerr := os.Create(path)
		if cerr == nil {
			_ = f.Close()
			return true
		}
	}
	return false
}

// --- middleware ---
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// --- модели/типы ---
type Product struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Specs string `json:"specs"`
	Price int    `json:"price"`
	Image string `json:"image"`
}

type authReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// --- handlers ---
func handleProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rows, err := db.Query("SELECT id, title, specs, price, image FROM products")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var list []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Title, &p.Specs, &p.Price, &p.Image); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, p)
	}
	writeJSON(w, list)
}

func handleProductByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/products/")
	row := db.QueryRow("SELECT id, title, specs, price, image FROM products WHERE id=?", id)
	var p Product
	if err := row.Scan(&p.ID, &p.Title, &p.Specs, &p.Price, &p.Image); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, p)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req authReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if req.Email == "" || len(req.Password) < 4 {
		http.Error(w, "invalid data", http.StatusBadRequest)
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	_, err := db.Exec("INSERT INTO users(email, password_hash) VALUES(?, ?)", req.Email, string(hash))
	if err != nil {
		http.Error(w, "email exists", http.StatusConflict)
		return
	}
	writeJSON(w, map[string]string{"status": "registered"})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req authReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	var hash string
	row := db.QueryRow("SELECT password_hash FROM users WHERE email=?", req.Email)
	if err := row.Scan(&hash); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	// простая cookie-сессия для демо
	cookie := &http.Cookie{
		Name:     "session",
		Value:    req.Email,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	writeJSON(w, map[string]string{"status": "ok"})
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	email := ""
	if c, err := r.Cookie("session"); err == nil {
		email = c.Value
	}
	if email == "" {
		http.Error(w, "auth required", http.StatusUnauthorized)
		return
	}
	// Демонстрация использования авторизации: заказ не сохраняем (лимит 2 таблиц)
	writeJSON(w, map[string]string{"status": "order accepted"})
}

// --- утилиты ---
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func readFile(path string) ([]byte, error) {
	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, errors.New("cannot read file")
	}
	return b, nil
}

// Сидим стартовые товары только если таблица пустая
func seedProductsIfEmpty() {
	var cnt int
	_ = db.QueryRow("SELECT COUNT(*) FROM products").Scan(&cnt)
	if cnt > 0 {
		return
	}
	_, _ = db.Exec(`
		INSERT INTO products (title, specs, price, image) VALUES
		('Игровой ПК Nitro X', 'Ryzen 5 5600, 16GB RAM, RTX 3060, SSD 512GB', 89990, '/public/img/pc1.jpg'),
		('Игровой ПК Nitro 2X', 'Intel i7, 32GB RAM, RTX 3060, SSD 1024GB', 134990, '/public/img/pc2.jpg'),
		('Ультра ПК Creator', 'Ryzen 7 7800X, 32GB RAM, RTX 4070, SSD 1TB', 169990, '/public/img/pc3.jpg'),
		('Компактный ПК Mini', 'Intel N100, 8GB RAM, SSD 256GB', 25990, '/public/img/pc4.jpg');
	`)
}
