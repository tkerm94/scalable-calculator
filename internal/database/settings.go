package database

func GetSettings(login string) (map[string]int, error) {
	rowsRs, err := DB.Query("SELECT * FROM Settings WHERE login = $1", login)
	if err != nil {
		return nil, err
	}
	defer rowsRs.Close()
	var add, sub, mult, div int
	for rowsRs.Next() {
		if err := rowsRs.Scan(&login, &add, &sub, &mult, &div); err != nil {
			return nil, err
		}
	}
	values := map[string]int{
		"add":  add,
		"sub":  sub,
		"mult": mult,
		"div":  div,
	}
	return values, nil
}
