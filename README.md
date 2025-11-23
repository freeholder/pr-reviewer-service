# PR Reviewer Service

Сервис для автоматического назначения ревьюеров на Pull Request'ы внутри команды.  
Реализация тестового задания для стажировки Backend (Авито, осенняя волна 2025).

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

После старта сервис доступен по адресу: http://localhost:8080

Быстрый smoke-тест:
```bash
curl http://localhost:8080/health
```
### Запуск в фоне + нагрузочный тест k6
```bash
make docker-up-app-db

make k6-test
```
Команды определены в Makefile.

## Конфигурация

Конфигурация читается из переменных окружения (или файла .env при локальном запуске).

Используются:

*APP_HTTP_PORT* — порт HTTP-сервера (по умолчанию 8080);

*APP_DB_DSN* — строка подключения к PostgreSQL, например:
```bash
postgres://pr_user:pr_password@db:5432/pr_db?sslmode=disable
```

В docker-compose.yml эти переменные уже выставлены для сервиса app.

Файл .env в git не коммитится – в репозитории лежит только .env.example.
