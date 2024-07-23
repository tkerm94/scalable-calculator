package database

import (
	"slices"
)

type Agent struct {
	ID         int
	Last_Seen  string
	Status     string
	Goroutines int
	Dead_Time  string
}

func GetAgentsFromDB() ([]Agent, error) {
	rowsRs, err := DB.Query("SELECT * FROM Agents")
	if err != nil {
		return nil, err
	}
	defer rowsRs.Close()
	agents := make([]Agent, 0)
	for rowsRs.Next() {
		agent := Agent{}
		if err := rowsRs.Scan(&agent.ID, &agent.Last_Seen, &agent.Status, &agent.Goroutines, &agent.Dead_Time); err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	slices.Reverse(agents)
	return agents, nil
}

func InsertAgentIntoDB(agent *Agent) error {
	query := `INSERT INTO Agents(last_seen, status, goroutines, dead_time) VALUES($1, $2, $3, $4)`
	_, err := DB.Exec(query, agent.Last_Seen, agent.Status, agent.Goroutines, agent.Dead_Time)
	return err
}
