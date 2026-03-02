# HwWach Go Skill

## Описание

Специализированный навык для работы с Go-проектом **HwWach** — веб-приложение для управления устройствами и фотографиями.

## Архитектура проекта

```
cmd/
  └── main.go              # Точка входа
internal/
  ├── app/                 # Инициализация приложения
  ├── config/              # Конфигурация (Viper)
  ├── handlers/            # HTTP-обработчики (Interface Layer)
  ├── middleware/          # JWT-аутентификация
  ├── models/              # Доменные модели (GORM)
  ├── repository/          # Доступ к данным (Data Access Layer)
  ├── routes/              # Маршрутизация (Gin)
  ├── services/            # Бизнес-логика (Business Layer)
  ├── storage/             # MinIO-хранилище
  └── utils/               # Утилиты
```

## Слои архитектуры

| Слой | Пакет | Ответственность |
|------|-------|-----------------|
| Presentation | `routes`, `middleware` | Маршрутизация, JWT |
| Interface | `handlers` | Обработка HTTP-запросов |
| Business | `services` | Бизнес-логика |
| Data Access | `repository` | Работа с БД |
| Domain | `models` | Доменные модели |

## Технологии

- **Go 1.24.5**
- **Gin-Gonic** — HTTP-фреймворк
- **GORM** — ORM для PostgreSQL
- **PostgreSQL** — основная БД
- **MinIO** — объектное хранилище (фотографии)
- **JWT** — аутентификация
- **Viper** — конфигурация (TOML + env)

## Структура данных

### Модели

```go
User     // id, login, password_hash, full_name, phone
Device   // устройство пользователя
Photo    // фотографии в MinIO
Request  // запросы на доступ/операции
```

## API Endpoints

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | `/auth/login` | Вход | ❌ |
| POST | `/auth/register` | Регистрация | ❌ |
| POST | `/auth/logout` | Выход | ❌ |
| POST | `/auth/refresh` | Refresh токена | ❌ |
| GET | `/users` | Профиль пользователя | ✅ |
| PATCH | `/users` | Обновление профиля | ✅ |
| PATCH | `/users/password` | Смена пароля | ✅ |
| GET | `/users/devices` | Список устройств | ✅ |
| GET | `/devices` | Все устройства | ✅ |
| GET | `/devices/:id` | Устройство по ID | ✅ |
| GET | `/devices/:id/photos` | Фото устройства | ✅ |
| POST | `/photos/upload` | Загрузка фото | ✅ |
| GET | `/photos/user` | Фото пользователя | ✅ |
| DELETE | `/photos/:id` | Удаление фото | ✅ |
| POST | `/requests` | Создать запрос | ✅ |
| GET | `/requests/:id` | Запрос по ID | ✅ |
| DELETE | `/requests/:id` | Удалить запрос | ✅ |

## Конфигурация

Файл: `config.toml` (опционально) или переменные окружения.

```toml
ServerAddress = ":8080"
DatabaseDSN = "host=localhost user=app dbname=app sslmode=disable password=secret"
JWTSecret = "your_jwt_secret_here"
MinioEndpoint = "play.min.io:9000"
MinioAccessKey = "..."
MinioSecretKey = "..."
MinioUseSSL = true
MinioBucket = "photos"
```

## Команды разработки

```bash
# Сборка
go build -o hwwach.exe ./cmd/main.go

# Запуск
go run ./cmd/main.go

# Тесты
go test ./...

# Форматирование
go fmt ./...

# Линтер (требуется golangci-lint)
golangci-lint run

# Модули
go mod tidy
go mod vendor
```

## Code Style

### Именование
- Пакеты: `snake_case` (нижний регистр)
- Типы/структуры: `PascalCase` (User, DeviceHandler)
- Функции/методы: `PascalCase` (экспорт), `camelCase` (внутренние)
- Переменные: `camelCase`
- Константы: `PascalCase` или `SCREAMING_SNAKE_CASE`

### Обработка ошибок
```go
// Всегда проверять ошибки
if err != nil {
    return nil, fmt.Errorf("описание: %w", err)
}

// Использовать wrapping для контекста
```

### Handler pattern
```go
type UserHandler interface {
    GetProfile(c *gin.Context)
    UpdateProfile(c *gin.Context)
}

// Реализация
type userHandler struct {
    userSvc services.UserService
}

func (h *userHandler) GetProfile(c *gin.Context) {
    // Извлечь userID из контекста (JWT)
    // Вызвать сервис
    // Вернуть JSON-ответ
}
```

### Service pattern
```go
type UserService interface {
    GetUserByID(id uint) (*models.User, error)
}

type userService struct {
    repo repository.UserRepository
}
```

### Repository pattern
```go
type UserRepository interface {
    Create(user *models.User) error
    FindByID(id uint) (*models.User, error)
    Update(user *models.User) error
}
```

## Best Practices проекта

1. **Зависимости**: импорты через модуль `HwWach/...`
2. **Контекст**: передавать `context.Context` в репозитории
3. **Валидация**: использовать `validator/v10` для входных данных
4. **Логирование**: стандартный `log` или `zap` (если добавлен)
5. **Миграции**: `db.AutoMigrate()` в `app.go`
6. **JWT Claims**: извлекать userID из токена в middleware

## Известные ограничения

- Проект на начальной стадии — многие обработчики требуют реализации
- Нет тестов (требуется добавить)
- Нет swagger-документации

## Полезные ссылки

- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [MinIO Go SDK](https://min.io/docs/minio/linux/developers/go/API.html)
