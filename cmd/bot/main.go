package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/zhashkevych/go-pocket-sdk"
	"tg-giga/pkg/config"
	"tg-giga/pkg/repository"
	"tg-giga/pkg/repository/boltdb"
	"tg-giga/pkg/server"
	"tg-giga/pkg/telegram"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatal("ERROR [CONFIG_INIT]", err)
	}

	fmt.Println("CONFIG: ", cfg)

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	pocketClient, err := pocket.NewClient(cfg.PocketConsumerKey)
	if err != nil {
		log.Fatal("ERROR [POCKET_INIT]", err)
	}

	db, err := initDB(cfg)
	if err != nil {
		log.Fatal("ERROR [DATABASE]", err)
	}

	tokenRepository := boltdb.NewTokenRepository(db)

	telegramBot := telegram.NewBot(
		bot,
		pocketClient,
		tokenRepository,
		cfg.AuthServerURL,
		cfg.Messages,
	)

	authorizationServer := server.NewAuthorizationServer(pocketClient, tokenRepository, cfg.TelegramBotURL)

	// Операция Start является блокирующей, потому что внутри неё идёт запись в канал, поэтому нужна обёртка из горутины
	go func() {
		if err := telegramBot.Start(); err != nil {
			log.Fatal("ERROR [START]", err)
		}
	}()

	if err := authorizationServer.Start(); err != nil {
		log.Fatal("ERROR [AUTH_START]", err)
	}
}

func initDB(cfg *config.Config) (*bolt.DB, error) {
	db, err := bolt.Open(cfg.DBPath, 0600, nil)
	if err != nil {
		fmt.Println("DB Init [ERROR]", err)
		return nil, err
	}

	// Создаст Buckets (таблицы)
	if err := db.Batch(func(tx *bolt.Tx) error {
		_, atError := tx.CreateBucketIfNotExists([]byte(repository.AccessTokens))
		if atError != nil {
			return atError
		}
		_, rtError := tx.CreateBucketIfNotExists([]byte(repository.RequestTokens))
		if rtError != nil {
			return rtError
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return db, nil
}
