package orchestrator

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"

	"github.com/tkerm94/scalable-calculator/internal/database"
)

func CheckPassword(password string) bool {
	matches := []bool{
		regexp.MustCompile(`^.{8,32}$`).MatchString(password),
		regexp.MustCompile(`[0-9]`).MatchString(password),
		regexp.MustCompile(`[a-z]`).MatchString(password),
		regexp.MustCompile(`[A-Z]`).MatchString(password),
	}
	for _, match := range matches {
		if !match {
			return false
		}
	}
	return true
}

func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/register" {
		http.NotFound(w, r)
		return
	}
	if r.Method == "POST" {
		login := r.PostFormValue("login")
		password := r.PostFormValue("password")
		if !CheckPassword(password) || login == "" {
			addr := "/register?error=Login+or+password+invalid."
			http.Redirect(w, r, addr, 303)
			return
		}
		exists, err := database.CheckIfUserExists(login)
		if err != nil {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if exists {
			addr := "/register?error=User+with+this+login+already+exists."
			http.Redirect(w, r, addr, 303)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		password = string(hash)
		users, err := database.GetUsersFromDB()
		if err != nil {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		user := database.User{
			ID:       len(users) + 1,
			Login:    login,
			Password: password,
		}
		if err := database.InsertUserIntoDB(&user); err != nil {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		GenerateAndSetToken(w, login)
		http.Redirect(w, r, "/calculator", 303)
		return
	}
	data := map[string]string{
		"Title": "Register",
		"Login": r.URL.Query().Get("user_login"),
		"Error": r.URL.Query().Get("error"),
	}
	tmpl, ok := Templates["register.html"]
	if !ok {
		c.Logger.Println("Error occured while providing register.html template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "register.html", data)
}

func (c *Controller) Register_api(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/register" {
		JSONResp(w, "Not Found", http.StatusNotFound)
		return
	}
	if r.Method != "POST" {
		JSONResp(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	validator := validator.New(validator.WithRequiredStructEnabled())
	var user database.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		JSONResp(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := validator.Struct(user); err != nil {
		JSONResp(w, "Bad Request", http.StatusBadRequest)
		return
	}
	exists, err := database.CheckIfUserExists(user.Login)
	if err != nil {
		c.Logger.Println(err)
		JSONResp(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if exists {
		JSONResp(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if !CheckPassword(user.Password) {
		JSONResp(w, "Bad Request", http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Logger.Println(err)
		JSONResp(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	user.Password = string(hash)
	users, err := database.GetUsersFromDB()
	if err != nil {
		c.Logger.Println(err)
		JSONResp(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	user.ID = len(users) + 1
	if err := database.InsertUserIntoDB(&user); err != nil {
		c.Logger.Println(err)
		JSONResp(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	JSONResp(w, "OK", http.StatusOK)
}
