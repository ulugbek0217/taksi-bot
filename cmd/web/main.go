package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/jackc/pgx/v5"
	"github.com/ulugbek0217/taksi-bot/pkg/handlers"
	hp "github.com/ulugbek0217/taksi-bot/pkg/helpers"
)

func main() {
	hp.LoadEnv("config/.env")

	db, err := pgx.Connect(context.Background(), os.Getenv("DB_PATH"))
	if err != nil {
		panic(err)
	}
	defer func(db *pgx.Conn, ctx context.Context) {
		err := db.Close(ctx)
		if err != nil {
			panic(err)
		}
	}(db, context.Background())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	admin_id, err := strconv.ParseInt(os.Getenv("ADMIN_ID"), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("cannot get admin id from env: %v", err))
	}
	// Configuring handlers
	app := &handlers.Handlers{
		DB:    db,
		Admin: admin_id,
	}

	opts := []bot.Option{
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, app.Start),
		bot.WithMessageTextHandler("🏠 Bosh sahifa", bot.MatchTypeExact, app.Start),
		bot.WithMessageTextHandler("/getcount", bot.MatchTypeExact, app.BotUsersQuantity),
		bot.WithMessageTextHandler("📍", bot.MatchTypePrefix, app.Direction),
		bot.WithMessageTextHandler("📦Pochta bor", bot.MatchTypeExact, app.Package),
		bot.WithDefaultHandler(app.Order),
	}

    bot_token := os.Getenv("TOKEN")
	b, err := bot.New(bot_token, opts...)
	if err != nil {
		fmt.Printf("couldn't create bot instance: %v", err)
	}

	fmt.Println("Starting the bot ...")
    resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/deleteWebhook?drop_pending_updates=1", bot_token))
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	b.Start(ctx)
}
