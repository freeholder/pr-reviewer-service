# PR Reviewer Service

Сервис для автоматического назначения ревьюеров на Pull Request'ы внутри команды. Реализация тестового задания для стажировки Backend (Авито, осенняя волна 2025).

- **Язык**: Go
- **БД**: PostgreSQL
- **Архитектура**: слои domain → repository → service → transport (HTTP)
- **Сборка и запуск**: `docker-compose up`
- **Порт сервиса**: 8080

## Как запустить

###  Через docker-compose
Требуется установленный Docker и docker-compose.

```bash
git clone https://github.com/freeholder/pr-reviewer-service
cd pr-reviewer-service

docker-compose up
```
При этом:
- поднимется база PostgreSQL 14-alpine;
- соберётся и запустится Go-сервис (app);
- автоматически применятся миграции (через goose при старте приложения).

После старта сервис доступен по адресу: ```http://localhost:8080```

Быстрый smoke-тест:
```bash
curl http://localhost:8080/health
```
### Запуск на фоне + нагрузочный тест k6
```bash
make docker-up-app-db

make k6-test
```
Команды определены в Makefile.


## Конфигурация

Конфигурация читается из переменных окружения (или файла .env при локальном запуске). 
Используются:

```APP_HTTP_PORT``` — порт HTTP-сервера (по умолчанию 8080);

```APP_DB_DSN``` — строка подключения к PostgreSQL, например:


```postgres://pr_user:pr_password@db:5432/pr_db?sslmode=disable```


В docker-compose.yml эти переменные уже выставлены для сервиса app. Файл .env в git не коммитится – в репозитории лежит только .env.example.

## Основные эндпоинты
Полное описание API — в ```openapi/openapi.yml```
(спецификация из задания, реализация соответствует ей).

Ниже — краткий обзор основных эндпоинтов.

### Teams
#### ```POST /team/add``` — cоздать команду с участниками (пользователи создаются или обновляются).

Пример запроса:
```json
{
  "team_name": "backend",
  "members": [
    { "user_id": "u1", "username": "Alice",   "is_active": true },
    { "user_id": "u2", "username": "Bob",     "is_active": true },
    { "user_id": "u3", "username": "Charlie", "is_active": true }
  ]
}
```

Пример ответа ```201 Created```:
```json
{
  "team": {
    "team_name": "backend",
    "members": [
      { "user_id": "u1", "username": "Alice",   "is_active": true },
      { "user_id": "u2", "username": "Bob",     "is_active": true },
      { "user_id": "u3", "username": "Charlie", "is_active": true }
    ]
  }
}
```

Возможные ошибки:

```400 TEAM_EXISTS``` — команда с таким team_name уже существует;

```400 BAD_REQUEST``` — не прошла валидация тела запроса.

#### ```GET /team/get?team_name=<name>``` — получить команду с участниками:

Пример ответа ```200 OK```:
```json
{
  "team_name": "backend",
  "members": [
    { "user_id": "u1", "username": "Alice", "is_active": true },
    { "user_id": "u2", "username": "Bob",   "is_active": true }
  ]
}
```
