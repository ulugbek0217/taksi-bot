
# TAXI BOT

A Telegram bot written in Go which takes orders from users for taxi. It may take passengers and packages to deliver. Language is Uzbek.


## How to run
Needed packages:
- Go 1.22.3 or newer version.
- Postgresql 10 or newer version.

Clone the repository from Github. Open terminal inside the repository. 
Fill the `config/.env` file with corresponding data. Run these commands to run the bot.

```bash
  go mod tidy
  go run cmd/web/main.go
```
    
