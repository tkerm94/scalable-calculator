package orchestrator

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/tkerm94/scalable-calculator/internal/database"
)

func (c *Controller) Settings(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/settings" {
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
	login, err := GetLoginFromToken(r)
	if err != nil {
		c.Logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if r.Method == "POST" {
		if err := ChangeSettings(login, r); err != nil && err.Error() != "Wrong syntax" {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings", 303)
		return
	}
	data := make(map[string]any)
	settings, err := database.GetSettings(login)
	for key, val := range settings {
		data[key] = val
	}
	if err != nil {
		c.Logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	agents, err := database.GetAgentsFromDB()
	if err != nil {
		c.Logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data["Agents"] = agents
	data["Title"] = "Settings"
	tmpl, ok := Templates["settings.html"]
	if !ok {
		c.Logger.Println("Error occured while providing settings.html template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "settings.html", data)
}

func ChangeSettings(login string, r *http.Request) error {
	add := r.PostFormValue("add")
	sub := r.PostFormValue("sub")
	mult := r.PostFormValue("mult")
	div := r.PostFormValue("div")
	values := []string{add, sub, mult, div}
	for _, value := range values {
		if !regexp.MustCompile(`^[1-9]{1}[0-9]*$`).MatchString(value) {
			return errors.New("Wrong syntax")
		}
	}
	query := `UPDATE Settings SET add = $1, sub = $2, mult = $3, div = $4 WHERE login = $5`
	_, err := database.DB.Exec(query, add, sub, mult, div, login)
	return err
}
