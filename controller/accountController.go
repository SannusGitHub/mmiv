package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type SessionData struct {
	Username string
	Expiry   time.Time
}

var Sessions = make(map[string]SessionData)

func GetCookie(r *http.Request, tokenName string) *http.Cookie {
	currentCookie, err := r.Cookie(tokenName)
	if err != nil {
		fmt.Println("cookie cannot be requested: " + tokenName)
		return nil
	}

	return currentCookie
}

func GetUsernameFromCookie(r *http.Request, tokenName string) string {
	cookie := GetCookie(r, tokenName)
	if cookie == nil {
		fmt.Println("No user from GetUserCookie, sorry!")
		return ""
	}

	session, ok := Sessions[cookie.Value]
	if !ok {
		return ""
	}

	return session.Username
}

func DoesUserMatchRank(r *http.Request, requiredRank string) bool {
	currentUser := GetUsernameFromCookie(r, "userSessionToken")
	userRankStr := GetUserRank(currentUser)

	userRank, err1 := strconv.Atoi(userRankStr)
	required, err2 := strconv.Atoi(requiredRank)

	if err1 != nil || err2 != nil {
		log.Printf("Rank conversion error: userRank=%s, requiredRank=%s\n", userRankStr, requiredRank)
		return false
	}

	// fmt.Printf("%s: %d >= %d\n", currentUser, userRank, required)
	return userRank >= required
}

func SetUserSessionCookie(w http.ResponseWriter, data UserData) {
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(86400 * time.Second)

	Sessions[sessionToken] = SessionData{
		Username: data.Username,
		Expiry:   expiresAt,
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "userSessionToken",
		Value:    sessionToken,
		Expires:  expiresAt,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

type UserData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Rank     string `json:"rank"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var data UserData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !CheckCredentials(data.Username, data.Password) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	fmt.Printf("User of %s has requested login successfully\n", data.Username)
	SetUserSessionCookie(w, data)

	// redirect to the actual website
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "success",
		"redirect": "/",
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// TODO: add in a session clear function here to prevent dead sessions from logout, maybe?

	http.SetCookie(w, &http.Cookie{
		Name:     "userSessionToken",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "logged_out",
	})
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "2") {
		fmt.Printf("Rank mismatch, invalid perms!\n")
		return
	}

	var data UserData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	hashedPassword, err := HashPassword(data.Password)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User Added\n")
	fmt.Printf("Username: %s\n", data.Username)
	fmt.Printf("Password: %s\n", hashedPassword)
	fmt.Printf("Rank: %s\n", data.Rank)

	WriteToSQL(`
		INSERT INTO USERS (username, password, rank)
		VALUES (?, ?, ?)
	`, data.Username, hashedPassword, data.Rank)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	if !DoesUserMatchRank(r, "2") {
		fmt.Printf("Rank mismatch, invalid perms!\n")
		return
	}

	var data UserData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	WriteToSQL(`
		DELETE FROM USERS WHERE username = ?
	`, data.Username)

	fmt.Printf("User of %s deleted successfully\n", data.Username)
}

func GetUserRank(username string) string {
	return QueryFromSQL(`
		SELECT rank FROM USERS WHERE username = ?
	`, username)
}

func CheckCredentials(username, password string) bool {
	passwordHash := QueryFromSQL(`
		SELECT password FROM USERS WHERE username = ?
	`, username)

	return VerifyPassword(password, passwordHash)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
