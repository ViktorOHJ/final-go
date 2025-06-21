package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var (
	DB *sql.DB
)

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(255) NOT NULL,
    comment TEXT,
    date TEXT NOT NULL,
    repeat VARCHAR(100)
);

CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

// Init инициализирует базу данных, создавая таблицы, если они не существуют.
func Init(dbFile string) error {
	install := false
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		install = true
	}

	database, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("ошибка при открытии БД: %w", err)
	}

	if err := database.Ping(); err != nil {
		return fmt.Errorf("ошибка при пинге БД: %w", err)
	}

	if install {
		_, err := database.Exec(schema)
		if err != nil {
			return fmt.Errorf("ошибка при инициализации схемы: %w", err)
		}
	}

	DB = database
	return nil
}
