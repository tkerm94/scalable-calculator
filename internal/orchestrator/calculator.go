package orchestrator

import (
	"net/http"

	"github.com/tkerm94/scalable-calculator/internal/database"
	messagebroker "github.com/tkerm94/scalable-calculator/internal/message_broker"
)

func (c *Controller) Calculator(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/calculator" {
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
		expr := r.PostFormValue("text")
		cl := &database.Calculation{Login: login, Expression: expr, Status: "in progress", Answer: 0}
		if err := database.InsertClIntoDB(cl); err != nil {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		cls, err := database.GetClsFromDB(login)
		if err != nil {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInsufficientStorage)
			return
		}
		cl.ID = cls[0].ID
		err = messagebroker.PublishCl(cl)
		if err != nil {
			c.Logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/calculator", 303)
		return
	}
	cls, err := database.GetClsFromDB(login)
	if err != nil {
		c.Logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"Calculations": cls,
		"Title":        "Calculator",
	}
	tmpl, ok := Templates["calculator.html"]
	if !ok {
		c.Logger.Println("Error occured while providing calculator.html template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "calculator.html", data)
}
