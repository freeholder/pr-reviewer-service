# PR Reviewer Service

Сервис для автоматического назначения ревьюеров на Pull Request'ы внутри команды. Реализация тестового задания для стажировки Backend (Авито, осенняя волна 2025).

- **Язык**: Go
- **БД**: PostgreSQL
- **Архитектура**: слои domain → repository → service → transport (HTTP)
- **Сборка и запуск**: `docker-compose up`
- **Порт сервиса**: 8080

## Запуск

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
- автоматически применятся миграции (через goose при старте приложения);
- применится нагрузочный тест k6.

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

**```POST /team/add```** — cоздать команду с участниками (пользователи создаются или обновляются)

Пример запроса:
```json
{
  "team_name": "backend",
  "members": [
    { "user_id": "u1", "username": "Egor",   "is_active": true },
    { "user_id": "u2", "username": "Andrei",  "is_active": true },
    { "user_id": "u3", "username": "Nikolai", "is_active": true },
    { "user_id": "u4", "username": "Vladimir", "is_active": true },
    { "user_id": "u5", "username": "Kirill", "is_active": true }
  ]
}
```

Пример ответа ```201 Created```:
```json
{
    "team": {
        "team_name": "backend",
        "members": [
            {
                "user_id": "u1",
                "username": "Egor",
                "is_active": true
            },
            {
                "user_id": "u2",
                "username": "Andrei",
                "is_active": true
            },
            {
                "user_id": "u3",
                "username": "Nikolai",
                "is_active": true
            },
            {
                "user_id": "u4",
                "username": "Vladimir",
                "is_active": true
            },
            {
                "user_id": "u5",
                "username": "Kirill",
                "is_active": true
            }
        ]
    }
}
```

Возможные ошибки:

```400 TEAM_EXISTS``` — команда с таким team_name уже существует;

```400 BAD_REQUEST``` — не прошла валидация тела запроса.

**```GET /team/get?team_name=<name>```**— получить команду с участниками.

Пример ответа ```200 OK```:
```json
{
    "team_name": "backend",
    "members": [
        {
            "user_id": "u1",
            "username": "Egor",
            "is_active": true
        },
        {
            "user_id": "u2",
            "username": "Andrei",
            "is_active": true
        },
        {
            "user_id": "u3",
            "username": "Nikolai",
            "is_active": true
        },
        {
            "user_id": "u4",
            "username": "Vladimir",
            "is_active": true
        },
        {
            "user_id": "u5",
            "username": "Kirill",
            "is_active": true
        }
    ]
}
```

### Users
**```POST /users/setIsActive```** — изменить флаг активности пользователя.

Пример запроса:
```json
{
  "user_id": "u2",
  "is_active": false
}
```

Пример ответа ```200 OK```:
```json

{
    "user": {
        "user_id": "u2",
        "username": "Andrei",
        "team_name": "backend",
        "is_active": false
    }
}
```
 **```GET /users/getReview?user_id=<id>```** — получить PR’ы, где пользователь назначен ревьювером.
 
 Пример ответа ```200 OK```:
```json
{
    "user_id": "u2",
    "pull_requests": []
}
```
### Pull Requests

 **```POST /pullRequest/create```** — cоздать PR и автоматически назначить до 2 ревьюверов из команды автора.

Пример запроса:
```json
{
  "pull_request_id": "pr-1001",
  "pull_request_name": "Add search",
  "author_id": "u1"
}
```

Пример ответа ```201 Created```:
```json

{
    "pr": {
        "pull_request_id": "pr-1001",
        "pull_request_name": "Add search",
        "author_id": "u1",
        "status": "OPEN",
        "assigned_reviewers": [
            "u3",
            "u5"
        ]
    }
}
```
**```POST /pullRequest/reassign```** — переназначить конкретного ревьювера на другого участника его команды.

Пример запроса:
```json
{
  "pull_request_id": "pr-1001",
  "old_user_id": "u3"
}
```

Пример ответа ```200 OK```:
```json
{
    "pr": {
        "pull_request_id": "pr-1001",
        "pull_request_name": "Add search",
        "author_id": "u1",
        "status": "OPEN",
        "assigned_reviewers": [
            "u5",
            "u4"
        ],
        "createdAt": "2025-11-24T05:59:25.64984Z"
    },
    "replaced_by": "u4"
}
```


**```POST /pullRequest/merge```** — идемпотентная операция merge.

Пример запроса:
```json
{
  "pull_request_id": "pr-1001"
}
```

Пример ответа ```200 OK```:
```json

{
    "pr": {
        "pull_request_id": "pr-1001",
        "pull_request_name": "Add search",
        "author_id": "u1",
        "status": "MERGED",
        "assigned_reviewers": [
            "u5",
            "u4"
        ],
        "createdAt": "2025-11-24T05:59:25.64984Z",
        "mergedAt": "2025-11-24T06:00:38.674579Z"
    }
}

```
Повторный вызов merge возвращает то же состояние MERGED.


Возможные ошибки:

```409 PR_MERGED``` — нельзя менять ревьюверов после MERGED;

```409 NOT_ASSIGNED``` — указанный пользователь не был ревьювером этого PR;

```409 NO_CANDIDATE``` — не найден активный кандидат для замены в команде.
### Статистика 
**```GET /stats/reviewers```** — cтатистика назначений по ревьюверам.

Пример ответа ```200 OK```:
```json
{
    "reviewer_stats": [
        {
            "user_id": "u4",
            "username": "Vladimir",
            "assigned_count": 1
        },
        {
            "user_id": "u5",
            "username": "Kirill",
            "assigned_count": 1
        },
        {
            "user_id": "u1",
            "username": "Egor",
            "assigned_count": 0
        },
        {
            "user_id": "u2",
            "username": "Andrei",
            "assigned_count": 0
        },
        {
            "user_id": "u3",
            "username": "Nikolai",
            "assigned_count": 0
        }
    ]
}
```

### Массовая деактивация
**```POST /team/bulkDeactivate```** — массовая деактивация пользователей команды и безопасная переназначаемость открытых PR.

Пример запроса:
```json
{
  "team_name": "backend",
  "user_ids": ["u4", "u5"]
}
```

Пример ответа ```200 OK```:
```json
{
    "team_name": "backend",
    "deactivated_user_ids": [
        "u4",
        "u5"
    ],
    "reassigned_count": 0,
    "not_reassigned": []
}
```
Здесь пользователи команды были деактивированы, но не переназначены, так как PR не открыт.

## Нагрузочное тестирование (k6)

В docker-compose.yml описан сервис k6, использующий образ grafana/k6.

Запуск:
```bash
# поднять app + db на фоне
make docker-up-app-db

# запустить k6
make k6-test

# или напрямую
docker-compose run --rm k6
```
Результаты:

- p95 ```http_req_duration``` — десятки миллисекунд;
- все ответы по ```/health``` и ```/users/getReview``` успешные;
- по SLI из задания (≤ 300 мс, ≥ 99.9% успешных ответов) сервис на тестовой нагрузке проходит.

## E2E-тестирование
Каталог e2e содержит e2e-тест, покрывающий типичный бизнес-сценарий.

Запуск:
```bash
go test ./e2e -count=1
```

## Линтер
Используется ```golangci-lint```.

Запуск:
```bash
# установка линтера в ./bin
make golangci-lint

# запуск линтера
make lint
```
