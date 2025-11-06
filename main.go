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

	_ "modernc.org/sqlite"
	"golang.org/x/crypto/bcrypt"
)

//go:embed public
var publicFS embed.FS

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", "file:pcshop.db?cache=shared&_pragma=busy_timeout=10000")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	// применим schema.sql (удобно при первом старте)
	if _, err := os.Stat("schema.sql"); err == nil {
		b, _ := os.ReadFile("schema.sql")
		if _, err := db.Exec(string(b)); err != nil {
			log.Println("schema apply warning:", err)
		}
	}

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/api/products", handleProducts)
	mux.HandleFunc("/api/products/", handleProductByID)
	mux.HandleFunc("/api/register", handleRegister)
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/order", handleOrder)

	// статика из embed (корень = /)
	fs := http.FS(publicFS)
	mux.Handle("/", http.FileServer(fs))

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           logging(cors(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Println("PC Shop running on http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}

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

func handleProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	rows, err := db.Query("SELECT id, title, specs, price, image FROM products")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var list []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Title, &p.Specs, &p.Price, &p.Image); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		list = append(list, p)
	}
	writeJSON(w, list)
}

func handleProductByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/products/")
	row := db.QueryRow("SELECT id, title, specs, price, image FROM products WHERE id=?", id)
	var p Product
	if err := row.Scan(&p.ID, &p.Title, &p.Specs, &p.Price, &p.Image); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	writeJSON(w, p)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req authReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	if req.Email == "" || len(req.Password) < 4 {
		http.Error(w, "invalid data", 400)
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	_, err := db.Exec("INSERT INTO users(email, password_hash) VALUES(?, ?)", req.Email, string(hash))
	if err != nil {
		http.Error(w, "email exists", 409)
		return
	}
	writeJSON(w, map[string]string{"status": "registered"})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req authReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	var hash string
	row := db.QueryRow("SELECT password_hash FROM users WHERE email=?", req.Email)
	if err := row.Scan(&hash); err != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}
	cookie := &http.Cookie{Name: "session", Value: req.Email, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode}
	http.SetCookie(w, cookie)
	writeJSON(w, map[string]string{"status": "ok"})
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	email := ""
	if c, err := r.Cookie("session"); err == nil {
		email = c.Value
	}
	if email == "" {
		http.Error(w, "auth required", 401)
		return
	}
	writeJSON(w, map[string]string{"status": "order accepted"})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func readFile(path string) ([]byte, error) {
	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, errors.New("cannot read file")
	}
	return b, nil
}
