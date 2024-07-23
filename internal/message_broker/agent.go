package messagebroker

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/streadway/amqp"
	"github.com/tkerm94/scalable-calculator/internal/agent"
	"github.com/tkerm94/scalable-calculator/internal/database"
)

func ConnectAgentToRMQ(id int, numGoroutines int) {
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
			"calculations_queue",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Fatalln(err)
		}
		q2, err := ch.QueueDeclare(
			"heartbeats_queue",
			false,
			false,
			false,
			false,
			nil,
		)
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
		go func() {
			for {
				body := strconv.Itoa(id)
				err = ch.Publish(
					"",
					q2.Name,
					false,
					false,
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(body),
					})
				if err != nil {
					return
				}
				time.Sleep(10 * time.Second)
			}
		}()
		log.Println("[*] Waiting for messages")
	receiving:
		for {
			select {
			case err = <-notify:
				log.Printf("An error occured while interacting with the broker, retrying in 30 seconds: %s", err)
				time.Sleep(30 * time.Second)
				break receiving
			case delivery := <-deliveryChan:
				var cl database.Calculation
				if err := json.Unmarshal(delivery.Body, &cl); err != nil {
					log.Println("Received invalid message")
					continue
				}
				if cl.Status != "in progress" {
					log.Println("Invalid data")
					continue
				}
				operations, err := database.GetSettings(cl.Login)
				if err != nil {
					log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
					time.Sleep(30 * time.Second)
					continue
				}
				log.Printf("Evaluating %s", cl.Expression)
				res, err := agent.EvaluateComplexExpression(cl.Expression, numGoroutines, operations)
				if err != nil {
					log.Printf("Error occured while parsing %s", cl.Expression)
					query := `UPDATE Calculations SET status = $1 WHERE id = $2`
					_, err := database.DB.Exec(query, "failed", cl.ID)
					if err != nil {
						log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
						time.Sleep(30 * time.Second)
						continue
					}
				} else {
					log.Printf("Done evaluating %s, result: %.2f", cl.Expression, res)
					query := `UPDATE Calculations SET status = $1, answer = $2 WHERE id = $3`
					_, err := database.DB.Exec(query, "ok", res, cl.ID)
					if err != nil {
						log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
						time.Sleep(30 * time.Second)
						continue
					}
				}
				delivery.Ack(false)
			default:
				time.Sleep(1 * time.Second)
			}
		}
	}
}
