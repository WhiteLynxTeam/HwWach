# HwWach - Docker развёртывание

## Быстрый старт

### Запуск всех сервисов

```bash
docker-compose up -d
```

Команды поднимут:
- **Приложение** на порту `8080`
- **PostgreSQL** на порту `5432`
- **MinIO** на порту `9000` (API) и `9001` (консоль)

### Остановка

```bash
docker-compose down
```

### Остановка с удалением данных

```bash
docker-compose down -v
```

## Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `SERVER_ADDRESS` | Адрес и порт сервера | `:8080` |
| `DATABASE_DSN` | Строка подключения к БД | - |
| `JWT_SECRET` | Секретный ключ для JWT | - |
| `MINIO_ENDPOINT` | Адрес MinIO | `minio:9000` |
| `MINIO_ACCESS_KEY` | Ключ доступа MinIO | - |
| `MINIO_SECRET_KEY` | Секретный ключ MinIO | - |
| `MINIO_USE_SSL` | Использовать SSL для MinIO | `false` |
| `MINIO_BUCKET` | Имя бакета MinIO | `photos` |

## Создание бакета MinIO

После запуска зайдите в консоль MinIO: http://localhost:9001

**Логин:** `hwwach_minio_key`  
**Пароль:** `hwwach_minio_secret`

Создайте бакет `photos` или измените `MINIO_BUCKET` в docker-compose.yml.

## Логи

```bash
# Все логи
docker-compose logs -f

# Лог приложения
docker-compose logs -f app

# Лог PostgreSQL
docker-compose logs -f postgres

# Лог MinIO
docker-compose logs -f minio
```

## Пересборка

```bash
docker-compose up -d --build
```

## Доступ к сервисам

| Сервис | URL |
|--------|-----|
| Приложение | http://localhost:8080 |
| PostgreSQL | localhost:5432 |
| MinIO API | http://localhost:9000 |
| MinIO Console | http://localhost:9001 |

## Безопасность

⚠️ **Важно:** Перед развёртыванием в production измените:
- `JWT_SECRET` на случайную строку
- Пароли для PostgreSQL и MinIO
