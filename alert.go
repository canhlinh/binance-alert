package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/naoina/go-stringutil"
)

type Alert struct {
	redisClient    *redis.Client
	httpClient     *http.Client
	bot            *tgbotapi.BotAPI
	alertChannelID int64
}

func NewAlert(redisHost string, botToken string, alertChannelID int64) *Alert {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	return &Alert{
		bot:            bot,
		alertChannelID: alertChannelID,
		redisClient:    redisClient,
		httpClient:     &http.Client{},
	}
}

func (alert *Alert) FetchNews() ([]string, error) {
	// Request the HTML page.
	res, err := alert.httpClient.Get(BinanceURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		return nil, errors.New(res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	// Find the review items
	news := []string{}
	doc.Find(".article-list .article-list-item ").Each(func(i int, s *goquery.Selection) {
		news = append(news, s.Find("a").Text())
	})
	return news, nil
}

func (alert *Alert) Exist(news string) (bool, error) {
	key := stringutil.ToSnakeCase(news)

	r, err := alert.redisClient.Exists(key).Result()
	if err != nil {
		return false, err
	}

	return r == 1, nil
}

func (alert *Alert) SaveNews(news string) error {
	key := stringutil.ToSnakeCase(news)
	return alert.redisClient.Set(key, true, 0).Err()
}

func (alert *Alert) Notify(news string) error {
	msg := tgbotapi.NewMessage(alert.alertChannelID, news)
	msg.ParseMode = "markdown"
	_, err := alert.bot.Send(msg)
	if err != nil {
		return err
	}

	if err := alert.SaveNews(news); err != nil {
		return err
	}

	return nil
}
