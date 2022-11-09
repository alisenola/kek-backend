package alert

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"

	alertDB "kek-backend/internal/alert/database"
	"kek-backend/internal/uniswap"

	"github.com/appleboy/go-fcm"
	"github.com/robfig/cron/v3"
)

func sendMessage(title string, body string, token string) {
	// Create the message to be sent.
	msg := &fcm.Message{
		To: token,
		Data: map[string]interface{}{
			"foo": "bar",
		},
		Notification: &fcm.Notification{
			Title: title,
			Body:  body,
		},
	}

	// Create a FCM client to send the message.
	client, err := fcm.NewClient("AAAAlSnRveU:APA91bF_XWeThMJnZuUGUyQ5wIBBBRyqGfryJ818ItRFUcJg0HubP6ekcNw0FF-ebQMHFZwva2wfEBIViv9qTh7QTeafiyHk8BWPgdE-j3DQEe2orVHpyayxF7DOyOujlarj2_SEhIr_")
	if err != nil {
		log.Fatalln(err)
	}

	// Send the message and receive the response without retries.
	response, err := client.Send(msg)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%#v\n", response)
}

func StartCron(db alertDB.AlertDB) {
	c := cron.New(cron.WithSeconds())
	c.AddFunc("@every 5s", func() {
		criteria := alertDB.IterateAlertCriteria{
			Account: 1,
			Offset:  0,
			Limit:   1,
		}
		alerts, _, err := db.FindAlertsWithoutContext(criteria)
		if err != nil {
			return
		}

		var wg sync.WaitGroup
		wg.Add(2)

		ch := make(chan int)

		var ethPrice float64
		var tokenPrice float64

		go func() {
			c1 := make(chan string, 1)
			query1 := uniswap.QueryBundles()
			uniswap.Request(query1, c1)

			msg1 := <-c1
			var bundles uniswap.Bundles
			json.Unmarshal([]byte(msg1), &bundles)
			ethPrice, _ = strconv.ParseFloat(bundles.Data.Bundles[0].EthPrice, 64)
			ch <- 1
			wg.Done()
		}()

		go func() {
			<-ch

			for _, alert := range alerts {
				c2 := make(chan string, 1)
				query2 := uniswap.QueryToken(alert.PairAddress)
				uniswap.Request(query2, c2)

				msg2 := <-c2
				var tokens uniswap.Tokens
				json.Unmarshal([]byte(msg2), &tokens)
				tokenPrice, _ = strconv.ParseFloat(tokens.Data.Tokens[0].DerivedETH, 64)
				go sendMessage(alert.Title, alert.Body, alert.Account.Token)
			}

			wg.Done()
		}()

		wg.Wait()
		fmt.Println("$$$$$:   ", ethPrice*tokenPrice)
	})
	c.Start()
}
