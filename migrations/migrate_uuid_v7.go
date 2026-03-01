package migrations

import (
	"gorm.io/gorm"
)

// MigrateUUIDv7 убирает default gen_random_uuid() из таблиц, так как UUID теперь генерируется в Go-коде (v7)
// Это необходимо, потому что gen_random_uuid() создаёт UUID v4 (случайный), а мы используем v7 (с timestamp)
func MigrateUUIDv7(db *gorm.DB) error {
	// Проверяем, существует ли таблица photos
	var tableName string
	err := db.Raw(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_name = 'photos'
	`).Scan(&tableName).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Если таблица существует, убираем default gen_random_uuid()
	if tableName == "photos" {
		// Проверяем, есть ли default значение у колонки uuid
		var columnDefault string
		err := db.Raw(`
			SELECT column_default
			FROM information_schema.columns
			WHERE table_name = 'photos' AND column_name = 'uuid'
		`).Scan(&columnDefault).Error

		if err == nil && columnDefault != "" {
			// Убираем default значение
			if err := db.Exec(`ALTER TABLE photos ALTER COLUMN uuid DROP DEFAULT`).Error; err != nil {
				return err
			}
		}
	}

	// Проверяем таблицу devices
	err = db.Raw(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_name = 'devices'
	`).Scan(&tableName).Error

	if err == nil && tableName == "devices" {
		var columnDefault string
		err := db.Raw(`
			SELECT column_default
			FROM information_schema.columns
			WHERE table_name = 'devices' AND column_name = 'uuid'
		`).Scan(&columnDefault).Error

		if err == nil && columnDefault != "" {
			if err := db.Exec(`ALTER TABLE devices ALTER COLUMN uuid DROP DEFAULT`).Error; err != nil {
				return err
			}
		}
	}

	// Проверяем таблицу requests
	err = db.Raw(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_name = 'requests'
	`).Scan(&tableName).Error

	if err == nil && tableName == "requests" {
		var columnDefault string
		err := db.Raw(`
			SELECT column_default
			FROM information_schema.columns
			WHERE table_name = 'requests' AND column_name = 'uuid'
		`).Scan(&columnDefault).Error

		if err == nil && columnDefault != "" {
			if err := db.Exec(`ALTER TABLE requests ALTER COLUMN uuid DROP DEFAULT`).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
