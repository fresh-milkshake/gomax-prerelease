package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Инкапсулирует подключение к SQLite‑базе для хранения сессионных данных MaxClient.
type DB struct {
	path string
	db   *sql.DB
}

// Открывает или создаёт файл базы session.db в указанной рабочей директории
// и выполняет миграции схемы.
func Open(workdir string) (*DB, error) {
	if err := os.MkdirAll(workdir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}
	path := filepath.Join(workdir, "session.db")
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	d := &DB{path: path, db: conn}
	if err := d.migrate(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return d, nil
}

// Применяет минимально необходимую схему для таблицы auth
// и гарантирует наличие одной служебной строки с device_id и токеном.
func (d *DB) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS auth (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  device_id TEXT NOT NULL,
  token TEXT,
  device_type TEXT
);
`
	if _, err := d.db.Exec(schema); err != nil {
		return err
	}
	var count int
	if err := d.db.QueryRow(`SELECT COUNT(*) FROM auth`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		devID := uuid.New().String()
		_, err := d.db.Exec(`INSERT INTO auth(device_id, token, device_type) VALUES(?,?,?)`,
			devID, "", "WEB",
		)
		return err
	}

	var existingID string
	if err := d.db.QueryRow(`SELECT device_id FROM auth LIMIT 1`).Scan(&existingID); err == nil {
		if _, err := uuid.Parse(existingID); err != nil {
			newUUID := uuid.New()
			_, err := d.db.Exec(`UPDATE auth SET device_id = ? WHERE id = (SELECT id FROM auth LIMIT 1)`,
				newUUID.String())
			if err != nil {
				return fmt.Errorf("failed to fix invalid device_id during migration: %w", err)
			}
		}
	}

	return nil
}

// Закрывает подключение к базе данных.
func (d *DB) Close() error {
	return d.db.Close()
}

// Возвращает сохранённый токен авторизации либо пустую строку, если токена нет.
func (d *DB) AuthToken() (string, error) {
	var token sql.NullString
	err := d.db.QueryRow(`SELECT token FROM auth LIMIT 1`).Scan(&token)
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", nil
	}
	return token.String, nil
}

// Возвращает идентификатор устройства из базы данных.
// Если device_id в базе невалидный, создает новый и обновляет базу.
func (d *DB) DeviceID() (uuid.UUID, error) {
	var id string
	if err := d.db.QueryRow(`SELECT device_id FROM auth LIMIT 1`).Scan(&id); err != nil {
		return uuid.UUID{}, err
	}

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		newUUID := uuid.New()
		_, updateErr := d.db.Exec(`UPDATE auth SET device_id = ? WHERE id = (SELECT id FROM auth LIMIT 1)`,
			newUUID.String())
		if updateErr != nil {
			return uuid.UUID{}, fmt.Errorf("failed to fix invalid device_id: %w", updateErr)
		}
		return newUUID, nil
	}

	return parsedUUID, nil
}

// Обновляет токен авторизации для указанного устройства в базе данных.
func (d *DB) UpdateToken(deviceID uuid.UUID, token string) error {
	_, err := d.db.Exec(`UPDATE auth SET token = ?, device_id = ? WHERE id = (SELECT id FROM auth LIMIT 1)`,
		token, deviceID.String(),
	)
	return err
}
