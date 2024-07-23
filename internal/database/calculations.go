package database

import "slices"

type Calculation struct {
	ID         int
	Login      string
	Expression string
	Status     string
	Answer     float64
}

func GetClsFromDB(login string) ([]Calculation, error) {
	rowsRs, err := DB.Query(`SELECT * FROM Calculations WHERE login = $1`, login)
	if err != nil {
		return nil, err
	}
	defer rowsRs.Close()
	cls := make([]Calculation, 0)
	for rowsRs.Next() {
		cl := Calculation{}
		if err := rowsRs.Scan(&cl.ID, &cl.Login, &cl.Expression, &cl.Status, &cl.Answer); err != nil {
			return nil, err
		}
		cls = append(cls, cl)
	}
	slices.Reverse(cls)
	return cls, nil
}

func InsertClIntoDB(c *Calculation) error {
	query := `INSERT INTO Calculations(login, expression, status, answer) VALUES($1, $2, $3, $4)`
	_, err := DB.Exec(query, c.Login, c.Expression, c.Status, c.Answer)
	return err
}
