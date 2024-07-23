package orchestrator

import (
	"net/http"
	"time"

	"github.com/tkerm94/scalable-calculator/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func (c *Controller) Profile(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/profile" {
		http.NotFound(w, r)
		return
	}
	if !ValidateToken(r) {
		http.Redirect(w, r, "/login", 303)
		return
	}
	err := RefreshToken(w, r)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if r.Method == "POST" {
		logout := r.PostFormValue("logout")
		if logout != "" {
			http.SetCookie(w, &http.Cookie{
				Name:    "token",
				Expires: time.Now(),
			})
			http.Redirect(w, r, "/login", 303)
			return
		}
		login, err := GetLoginFromToken(r)
		if err != nil {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		old_password := r.PostFormValue("old password")
		new_password := r.PostFormValue("new password")
		new_password_again := r.PostFormValue("new password again")
		corr_password, err := database.GetUserPassword(login)
		if err != nil {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if new_password != new_password_again {
			addr := "/profile?error=Passwords+do+not+match.+Try+again."
			http.Redirect(w, r, addr, 303)
			return
		}
		if !CheckPassword(new_password) {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(corr_password), []byte(old_password)); err != nil {
			addr := "/profile?error=Incorrect+password.+Try+again."
			http.Redirect(w, r, addr, 303)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(new_password), bcrypt.DefaultCost)
		if err != nil {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		password := string(hash)
		err = database.UpdateUserPassword(login, password)
		if err != nil {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/profile?message=Password+changed+successfully!", 303)
		return
	}
	login, err := GetLoginFromToken(r)
	if err != nil {
		c.Logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := map[string]string{
		"Title":   "Profile",
		"Login":   login,
		"Error":   r.URL.Query().Get("error"),
		"Message": r.URL.Query().Get("message"),
	}
	tmpl, ok := Templates["profile.html"]
	if !ok {
		c.Logger.Println("Error occured while providing profile.html template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "profile.html", data)
}
