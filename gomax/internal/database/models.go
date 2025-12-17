package database

// Хранит токен и deviceId, аналогично модели Auth в Python-версии.
type Auth struct {
	ID         int64  `db:"id"`
	DeviceID   string `db:"device_id"`
	Token      string `db:"token"`
	DeviceType string `db:"device_type"`
}
