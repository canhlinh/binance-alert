package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

const (
	BinanceURL = "https://www.binance.com/vn/support/sections/115000106672-Danh-s%C3%A1ch-ni%C3%AAm-y%E1%BA%BFt-m%E1%BB%9Bi"
)

func main() {
	redisHost := os.Getenv("REDIS_HOST")
	alertHook := os.Getenv("BOT_API_TOKEN")
	alertChannelID, _ := strconv.ParseInt(os.Getenv("ALERT_CHANNEL_ID"), 10, 64)

	ticker := time.NewTicker(time.Minute * 3)
	alert := NewAlert(redisHost, alertHook, alertChannelID)

	log.Println("App started")

LOOP:
	for {
		select {
		case <-ticker.C:
			newses, err := alert.FetchNews()
			if err != nil {
				log.Println(err)
				break LOOP
			}

			for _, news := range newses {
				if exist, err := alert.Exist(news); err != nil {
					log.Println(err)
					break LOOP
				} else if !exist {
					if err := alert.Notify(news); err != nil {
						log.Println(err)
						break LOOP
					}
				}
			}
		}
	}

	log.Println("App stopped")
}
