package migrations

import (
	"gorm.io/gorm"
)

// MigrateUintToUUID удаляет старые таблицы с uint ID, чтобы AutoMigrate создал их заново с UUID
// Это необходимо, потому что мы перешли с автоинкремента (uint) на UUID
// Функция безопасна: если таблицы уже с UUID — ничего не делает
func MigrateUintToUUID(db *gorm.DB) error {
	// Проверяем, существует ли таблица devices со старой схемой (id типа bigint/uint)
	// Если id уже uuid — миграция не нужна
	var columnType string
	err := db.Raw(`
		SELECT data_type
		FROM information_schema.columns
		WHERE table_name = 'devices' AND column_name = 'id'
	`).Scan(&columnType).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Если колонка уже uuid — миграция не нужна
	if columnType == "uuid" {
		return nil
	}

	// Таблицы нет или она с old-схемой — удаляем все зависимые таблицы
	// Порядок важен: сначала дочерние, потом родительские
	db.Exec(`DROP TABLE IF EXISTS requests CASCADE`)
	db.Exec(`DROP TABLE IF EXISTS photos CASCADE`)
	db.Exec(`DROP TABLE IF EXISTS devices CASCADE`)
	db.Exec(`DROP TABLE IF EXISTS device_photos CASCADE`)

	return nil
}
