package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/tkerm94/scalable-calculator/internal/database"
	messagebroker "github.com/tkerm94/scalable-calculator/internal/message_broker"
)

func main() {
	log.Println("Agent is starting...")
	numGoroutines := 5
	if len(os.Args) > 1 {
		val, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Println("An error occured while parsing arguments, starting with default settings")
		} else {
			numGoroutines = val
		}
	}
	log.Printf("Agent started with %d working goroutines", numGoroutines)
	database.InitDB()
	defer database.DB.Close()
	agent := &database.Agent{Last_Seen: time.Now().Local().Format("01/02/2006 15:04:05"), Status: "active", Goroutines: numGoroutines, Dead_Time: ""}
	for {
		if err := database.InsertAgentIntoDB(agent); err != nil {
			log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}
		agents, err := database.GetAgentsFromDB()
		if err != nil {
			log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}
		agent.ID = agents[0].ID
		break
	}
	var forever chan interface{}
	go messagebroker.ConnectAgentToRMQ(agent.ID, numGoroutines)
	<-forever
}
