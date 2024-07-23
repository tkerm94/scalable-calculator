package database

import (
	"slices"
)

type User struct {
	ID       int
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func GetUsersFromDB() ([]User, error) {
	rowsRs, err := DB.Query("SELECT * FROM Users")
	if err != nil {
		return nil, err
	}
	defer rowsRs.Close()
	users := make([]User, 0)
	for rowsRs.Next() {
		user := User{}
		if err := rowsRs.Scan(&user.ID, &user.Login, &user.Password); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	slices.Reverse(users)
	return users, nil
}

func GetUserPassword(login string) (string, error) {
	query := `SELECT password FROM Users WHERE login = $1`
	rowsRs, err := DB.Query(query, login)
	if err != nil {
		return "", err
	}
	defer rowsRs.Close()
	var password string
	for rowsRs.Next() {
		err = rowsRs.Scan(&password)
	}
	return password, err
}

func CheckIfUserExists(login string) (bool, error) {
	query := `SELECT id FROM Users WHERE login = $1`
	rowsRs, err := DB.Query(query, login)
	if err != nil {
		return false, err
	}
	defer rowsRs.Close()
	var id int
	for rowsRs.Next() {
		err = rowsRs.Scan(&id)
	}
	return id != 0, err
}

func InsertUserIntoDB(u *User) error {
	query := `INSERT INTO Users(id, login, password) VALUES($1, $2, $3)`
	_, err := DB.Exec(query, u.ID, u.Login, u.Password)
	query = `INSERT INTO Settings(login) VALUES($1)`
	_, err = DB.Exec(query, u.Login)
	return err
}

func UpdateUserPassword(login, password string) error {
	query := `UPDATE Users SET password = $1 WHERE login = $2`
	_, err := DB.Exec(query, password, login)
	return err
}
