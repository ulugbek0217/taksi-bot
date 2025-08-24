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

// GetSetting retrieves a setting value from the database
func GetSetting(ctx context.Context, db *pgx.Conn, key string) (string, error) {
	var value string
	err := db.QueryRow(ctx, "SELECT value FROM bot_settings WHERE key = $1", key).Scan(&value)
	return value, err
}

// SetSetting saves or updates a setting in the database
func SetSetting(ctx context.Context, db *pgx.Conn, key, value string) error {
	_, err := db.Exec(ctx, `
		INSERT INTO bot_settings (key, value) 
		VALUES ($1, $2) 
		ON CONFLICT (key) 
		DO UPDATE SET 
			value = $2, 
			updated_at = CURRENT_TIMESTAMP
	`, key, value)
	return err
}

// InitializeSettings ensures default settings exist in the database
func InitializeSettings(ctx context.Context, db *pgx.Conn) error {
	// Check if send_to_group setting exists, if not create it
	_, err := GetSetting(ctx, db, "send_to_group")
	if err != nil {
		// Setting doesn't exist, create it with default value
		err = SetSetting(ctx, db, "send_to_group", "false")
		if err != nil {
			return err
		}
	}
	return nil
}