package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/tkerm94/scalable-calculator/internal/database"
)

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		http.NotFound(w, r)
		return
	}
	if r.Method == "POST" {
		login := r.PostFormValue("login")
		password := r.PostFormValue("password")
		corr_password, err := database.GetUserPassword(login)
		if err != nil {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if corr_password == "" {
			addr := fmt.Sprintf("/login?error=Incorrect+login+or+password.+Try+again.&user_login=%s", login)
			http.Redirect(w, r, addr, 303)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(corr_password), []byte(password)); err != nil {
			addr := fmt.Sprintf("/login?error=Incorrect+login+or+password.+Try+again.&user_login=%s", login)
			http.Redirect(w, r, addr, 303)
			return
		}
		GenerateAndSetToken(w, login)
		http.Redirect(w, r, "/calculator", 303)
		return
	}
	data := map[string]string{
		"Title": "Log in",
		"Login": r.URL.Query().Get("user_login"),
		"Error": r.URL.Query().Get("error"),
	}
	tmpl, ok := Templates["login.html"]
	if !ok {
		c.Logger.Println("Error occured while providing login.html template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "login.html", data)
}

func (c *Controller) Login_api(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/login" {
		JSONResp(w, "Not Found", http.StatusNotFound)
		return
	}
	if r.Method != "POST" {
		JSONResp(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	validator := validator.New(validator.WithRequiredStructEnabled())
	var user_data database.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user_data); err != nil {
		JSONResp(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := validator.Struct(user_data); err != nil {
		JSONResp(w, "Bad Request", http.StatusBadRequest)
		return
	}
	password, err := database.GetUserPassword(user_data.Login)
	if err != nil {
		c.Logger.Println(err)
		JSONResp(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if password == "" {
		JSONResp(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(user_data.Password)); err != nil {
		JSONResp(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	expirationTime := time.Now().Add(12 * time.Hour)
	claims := &Claims{
		Login: user_data.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtKey))
	JSONResp(w, "OK", http.StatusOK, tokenString)
}
