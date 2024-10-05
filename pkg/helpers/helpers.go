package helpers

import (
	"github.com/jackc/pgx/v5"
	"context"
	"github.com/joho/godotenv"
)

// UserExists checks if the user exists
func UserExists(db *pgx.Conn, tgId int64) bool {
	var exists = true
	var data string
	err := db.QueryRow(context.Background(), "SELECT 1 FROM users WHERE user_id = $1", tgId).Scan(&data)
	if err != nil {
		exists = false
	}
	return exists
}

// RegMin registers the users' id in to the database
func RegMin(db *pgx.Conn, tgId int64) {
	_, err := db.Exec(context.Background(), "insert into users (user_id) values ($1)",  tgId)
	if err != nil {
		panic(err)
	}
}

// CountUsers returns the number of registered users
func CountUsers(ctx context.Context, db *pgx.Conn) (uint, error) {
	var count uint
	err := db.QueryRow(ctx, "select count(*) from users").Scan(&count)
	return count, err
}

// LoadEnv loads the environment variables
func LoadEnv(path string) error {
	err := godotenv.Overload(path)
	return err
}
