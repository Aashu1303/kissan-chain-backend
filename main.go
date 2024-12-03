package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"

	_ "modernc.org/sqlite" // SQLite driver
)

// User structure
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var db *sql.DB

// Initialize the database
func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./users.db")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Create users table if not exists
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

// Hash a password
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// Verify a password
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Sign-up handler
func signUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Insert user into the database
	query := "INSERT INTO users (username, password) VALUES (?, ?)"
	_, err = db.Exec(query, user.Username, hashedPassword)
	if err != nil {
		http.Error(w, "Error saving user", http.StatusInternalServerError)
		log.Printf("Failed to save user %s: %v", user.Username, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "User %s created successfully", user.Username)
}

// Login handler
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var req User
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Fetch user from the database
	query := "SELECT password FROM users WHERE username = ?"
	var hashedPassword string
	err = db.QueryRow(query, req.Username).Scan(&hashedPassword)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		log.Printf("No such user %s: %v", req.Username, err)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Query error for user %s: %v", req.Username, err)
		return
	}

	// Check password
	if !checkPasswordHash(req.Password, hashedPassword) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		log.Printf("Password mismatch for user %s", req.Username)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Welcome, %s!", req.Username)
}

func main() {
	// Initialize the database
	initDB()
	defer db.Close()

	// Define routes
	http.HandleFunc("/signup", signUpHandler)
	http.HandleFunc("/login", loginHandler)

	// Start the server
	fmt.Println("Server running on http://127.0.0.1:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
