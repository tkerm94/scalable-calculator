package messagebroker

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/streadway/amqp"

	"github.com/tkerm94/scalable-calculator/internal/database"
)

func PublishCl(c *database.Calculation) error {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"calculations_queue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	body, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func ConnectOrchestratorToRMQ() {
	for {
		conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err != nil {
			log.Printf("An error occured while interacting with the broker, retrying in 30 seconds: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}
		defer conn.Close()
		notify := conn.NotifyClose(make(chan *amqp.Error))
		ch, err := conn.Channel()
		if err != nil {
			log.Printf("An error occured while interacting with the broker, retrying in 30 seconds: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}
		defer ch.Close()
		q, err := ch.QueueDeclare(
			"heartbeats_queue",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Fatalln(err)
		}
		deliveryChan, err := ch.Consume(
			q.Name,
			"",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Printf("An error occured while interacting with the broker, retrying in 30 seconds: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}
		log.Println("Connected to RabbitMQ instance")
	receiving:
		for {
			select {
			case err = <-notify:
				log.Printf("An error occured while interacting with the broker, retrying in 30 seconds: %s", err)
				time.Sleep(30 * time.Second)
				break receiving
			case delivery := <-deliveryChan:
				id, err := strconv.Atoi(string(delivery.Body))
				if err != nil {
					log.Println("Invalid id format received from a heartbeat")
					continue
				}
				query := `UPDATE Agents SET last_seen = $1, status = $2 WHERE id = $3`
				_, err = database.DB.Exec(query, time.Now().Local().Format("01/02/2006 15:04:05"), "active", id)
				if err != nil {
					log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
					time.Sleep(30 * time.Second)
					continue
				}
				log.Printf("Received a heartbeat from agent %s", delivery.Body)
				delivery.Ack(false)
			default:
				agents, err := database.GetAgentsFromDB()
				if err != nil {
					log.Printf("An error occured while interacting with db, retrying in 30 seconds: %s", err)
					time.Sleep(30 * time.Second)
					continue
				}
				for _, agent := range agents {
					last_seen, err := time.ParseInLocation("01/02/2006 15:04:05", agent.Last_Seen, time.Local)
					if err != nil {
						log.Println("An error occured while parsing agent, deleting invalid item")
						query := `DELETE FROM Agents WHERE id = $1`
						_, err = database.DB.Exec(query, agent.ID)
						if err != nil {
							log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
							time.Sleep(30 * time.Second)
							continue
						}
						continue
					}
					if time.Since(last_seen) >= 30*time.Second && agent.Status == "active" {
						log.Printf("No heartbeats received from agent %d for a long time, changing status to inactive", agent.ID)
						query := `UPDATE Agents SET status = $1 WHERE id = $2`
						_, err = database.DB.Exec(query, "inactive", agent.ID)
						if err != nil {
							log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
							time.Sleep(1 * time.Minute)
							continue
						}
					} else if time.Since(last_seen) >= 1*time.Minute && agent.Status == "inactive" {
						log.Printf("No heartbeats received from agent %d for a very long time, changing status to dead", agent.ID)
						dead_time := time.Now().Local().Format("01/02/2006 15:04:05")
						query := `UPDATE Agents SET status = $1, dead_time = $2 WHERE id = $3`
						_, err = database.DB.Exec(query, "dead", dead_time, agent.ID)
						if err != nil {
							log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
							time.Sleep(1 * time.Minute)
							continue
						}
					} else if agent.Status == "dead" {
						dead_time, err := time.ParseInLocation("01/02/2006 15:04:05", agent.Dead_Time, time.Local)
						if time.Since(dead_time) >= 10*time.Minute {
							query := `DELETE FROM Agents WHERE id = $1`
							log.Println("Deleting dead agent")
							_, err = database.DB.Exec(query, agent.ID)
							if err != nil {
								log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
								time.Sleep(30 * time.Second)
								continue
							}
						}
					}
				}
				time.Sleep(1 * time.Second)
			}
		}
	}
}
