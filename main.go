package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"strings"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	CPU         string  `json:"cpu"`
	RAM         string  `json:"ram"`
	Storage     string  `json:"storage"`
	GPU         string  `json:"gpu"`
	Description string  `json:"description"`
}

type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type CartItem struct {
	ID        int     `json:"id"`
	UserID    int     `json:"user_id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Product   Product `json:"product"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var db *sql.DB

func main() {
	var err error

	db, err = sql.Open("sqlite3", "./pcshop.db")
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	if err := initDB(); err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/styles/",
		http.StripPrefix("/styles/",
			http.FileServer(http.Dir("styles")),
		),
	)
	mux.Handle("/scripts/",
		http.StripPrefix("/scripts/",
			http.FileServer(http.Dir("scripts")),
		),
	)
	mux.Handle("/images/",
		http.StripPrefix("/images/",
			http.FileServer(http.Dir("images")),
		),
	)

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/products", productsPageHandler)
	mux.HandleFunc("/contacts", contactsPageHandler)
	mux.HandleFunc("/profile", profilePageHandler)
	mux.HandleFunc("/cart", cartPageHandler)
	mux.HandleFunc("/api/products", productsHandler)
	mux.HandleFunc("/api/register", registerHandler)
	mux.HandleFunc("/api/login", loginHandler)
	mux.HandleFunc("/api/logout", logoutHandler)
	mux.HandleFunc("/api/user", userHandler)
	mux.HandleFunc("/api/cart", cartHandler)
	mux.HandleFunc("/api/cart/add", addToCartHandler)
	mux.HandleFunc("/api/cart/remove/", removeFromCartHandler)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Ошибка сервера:", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
    	port = "8080"
	}

	log.Println("Сервер запущен на порту", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
    	log.Fatal("Ошибка сервера:", err)
	}
}


func initDB() error {
	if err := runSQLFile("./db/schema.sql"); err != nil {
		return err
	}
	if err := runSQLFile("./db/seed.sql"); err != nil {
		return err
	}

	return nil
}

func runSQLFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	sqlText := string(data)
	if sqlText == "" {
		return nil
	}

	_, err = db.Exec(sqlText)
	return err
}


func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка при создании пользователя", http.StatusInternalServerError)
		return
	}

	result, err := db.Exec("INSERT INTO users (email, password, name) VALUES (?, ?, ?)",
		user.Email, string(hashedPassword), user.Name)
	if err != nil {
		http.Error(w, "Пользователь с таким email уже существует", http.StatusConflict)
		return
	}

	userID, _ := result.LastInsertId()
	user.ID = int(userID)
	user.Password = ""

	createSession(w, user.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	var user User
	var hashedPassword string
	err := db.QueryRow("SELECT id, email, password, name FROM users WHERE email = ?",
		loginReq.Email).Scan(&user.ID, &user.Email, &hashedPassword, &user.Name)
	if err != nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(loginReq.Password)); err != nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	createSession(w, user.ID)

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromSession(r)
	if userID == 0 {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	var user User
	err := db.QueryRow("SELECT id, email, name, created_at FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
 

func cartHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromSession(r)
	if userID == 0 {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	rows, err := db.Query(`
		SELECT ci.id, ci.user_id, ci.product_id, ci.quantity,
		       p.id, p.name, p.price, p.image, p.cpu, p.ram, p.storage, p.gpu, p.description
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.user_id = ?
	`, userID)
	if err != nil {
		http.Error(w, "Ошибка запроса к БД", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var cartItems []CartItem
	for rows.Next() {
		var item CartItem
		var product Product
		err := rows.Scan(&item.ID, &item.UserID, &item.ProductID, &item.Quantity,
			&product.ID, &product.Name, &product.Price, &product.Image, &product.CPU,
			&product.RAM, &product.Storage, &product.GPU, &product.Description)
		if err != nil {
			http.Error(w, "Ошибка чтения данных", http.StatusInternalServerError)
			return
		}
		item.Product = product
		cartItems = append(cartItems, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cartItems)
}

func addToCartHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromSession(r)
	if userID == 0 {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var item CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	var existingQuantity int
	err := db.QueryRow("SELECT quantity FROM cart_items WHERE user_id = ? AND product_id = ?",
		userID, item.ProductID).Scan(&existingQuantity)

	if err == sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO cart_items (user_id, product_id, quantity) VALUES (?, ?, ?)",
			userID, item.ProductID, 1)
	} else if err == nil {
		_, err = db.Exec("UPDATE cart_items SET quantity = quantity + 1 WHERE user_id = ? AND product_id = ?",
			userID, item.ProductID)
	}

	if err != nil {
		http.Error(w, "Ошибка при добавлении в корзину", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func removeFromCartHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromSession(r)
	if userID == 0 {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}

	itemID, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Неверный ID товара", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM cart_items WHERE id = ? AND user_id = ?", itemID, userID)
	if err != nil {
		http.Error(w, "Ошибка при удалении из корзины", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
 

func createSession(w http.ResponseWriter, userID int) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    fmt.Sprintf("%d", userID),
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

func getUserIdFromSession(r *http.Request) int {
	cookie, err := r.Cookie("session")
	if err != nil {
		return 0
	}

	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return 0
	}

	return userID
}


func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./html/index.html")
}

func productsPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./html/products.html")
}

func contactsPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./html/contacts.html")
}

func profilePageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./html/profile.html")
}


func cartPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./html/cart.html")
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query(`
		SELECT 
			id,
			name,
			price,
			image,
			cpu,
			ram,
			storage,
			gpu,
			description
		FROM products
	`)
	if err != nil {
		http.Error(w, "Ошибка запроса к БД", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Price,
			&p.Image,
			&p.CPU,
			&p.RAM,
			&p.Storage,
			&p.GPU,
			&p.Description,
		); err != nil {
			http.Error(w, "Ошибка чтения данных", http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(products); err != nil {
		http.Error(w, "Ошибка кодирования JSON", http.StatusInternalServerError)
		return
	}
}