package migrations

import (
	"gorm.io/gorm"
)

// MigratePhotoStatus добавляет поля status, file_size, file_name, content_type в таблицу photos
// Если таблица уже имеет новую схему — миграция не применяется
func MigratePhotoStatus(db *gorm.DB) error {
	// Проверяем, существует ли уже колонка status
	var columnName string
	err := db.Raw(`
		SELECT column_name
		FROM information_schema.columns
		WHERE table_name = 'photos' AND column_name = 'status'
	`).Scan(&columnName).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Если колонка status уже существует — миграция не нужна
	if columnName == "status" {
		return nil
	}

	// Добавляем новые поля в таблицу photos
	// Используем ALTER TABLE, так как таблица уже существует
	migrations := []string{
		`ALTER TABLE photos ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'pending'`,
		`ALTER TABLE photos ADD COLUMN IF NOT EXISTS file_size BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE photos ADD COLUMN IF NOT EXISTS file_name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE photos ADD COLUMN IF NOT EXISTS content_type VARCHAR(100) NOT NULL DEFAULT ''`,
	}

	for _, migration := range migrations {
		if err := db.Exec(migration).Error; err != nil {
			return err
		}
	}

	return nil
}

// MigratePhotoClientID добавляет поле client_id для оптимистичного UI
func MigratePhotoClientID(db *gorm.DB) error {
	// Проверяем, существует ли уже колонка client_id
	var columnName string
	err := db.Raw(`
		SELECT column_name
		FROM information_schema.columns
		WHERE table_name = 'photos' AND column_name = 'client_id'
	`).Scan(&columnName).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Если колонка уже существует — миграция не нужна
	if columnName == "client_id" {
		return nil
	}

	// Добавляем поле client_id
	if err := db.Exec(`ALTER TABLE photos ADD COLUMN client_id UUID`).Error; err != nil {
		return err
	}

	// Создаём уникальный индекс для client_id
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_photos_client_id ON photos (client_id)`).Error; err != nil {
		return err
	}

	return nil
}
