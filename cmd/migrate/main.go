package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pressly/goose/v3"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	var (
		command   = flag.String("command", "status", "Команда: up, down, status, redo, reset")
		dsn       = flag.String("dsn", "", "Database DSN (или переменная DATABASEDSN)")
		migrations = flag.String("dir", "migrations", "Путь к директории с миграциями")
	)
	flag.Parse()

	// Получаем DSN из флага или переменной окружения
	dbDSN := *dsn
	if dbDSN == "" {
		dbDSN = os.Getenv("DATABASEDSN")
	}
	if dbDSN == "" {
		log.Fatal("Требуется DSN: используйте флаг -dsn или переменную окружения DATABASEDSN")
	}

	// Подключаемся к БД
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		log.Fatalf("Ошибка ping БД: %v", err)
	}

	// Настраиваем goose
	goose.SetDialect("postgres")

	// Выполняем команду
	switch *command {
	case "up":
		if err := goose.Up(db, *migrations); err != nil {
			log.Fatalf("Ошибка выполнения up: %v", err)
		}
		fmt.Println("Миграции успешно применены")

	case "down":
		if err := goose.Down(db, *migrations); err != nil {
			log.Fatalf("Ошибка выполнения down: %v", err)
		}
		fmt.Println("Последняя миграция успешно откачена")

	case "redo":
		if err := goose.Redo(db, *migrations); err != nil {
			log.Fatalf("Ошибка выполнения redo: %v", err)
		}
		fmt.Println("Миграция успешно перезапущена")

	case "reset":
		if err := goose.Reset(db, *migrations); err != nil {
			log.Fatalf("Ошибка выполнения reset: %v", err)
		}
		fmt.Println("Все миграции успешно откачены")

	case "status":
		versions, err := goose.GetDBVersion(db)
		if err != nil {
			log.Fatalf("Ошибка получения версии: %v", err)
		}
		fmt.Printf("Текущая версия БД: %d\n", versions)

	default:
		log.Fatalf("Неизвестная команда: %s (доступные: up, down, status, redo, reset)", *command)
	}
}
