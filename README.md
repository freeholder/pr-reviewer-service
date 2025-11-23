# PR Reviewer Assignment Service

Сервис для автоматического назначения ревьюеров на Pull Request'ы внутри команды.  
Реализация тестового задания для стажировки Backend (Авито, осенняя волна 2025).

- **Язык**: Go
- **БД**: PostgreSQL
- **Архитектура**: слои domain → repository → service → transport (HTTP)
- **Сборка и запуск**: `docker-compose up`
- **Порт сервиса**: 8080

## 1. Как запустить

### Вариант 1. Основной (как в задании) — через docker-compose

Требуется установленный Docker и docker-compose.

```bash
git clone <url_репозитория>
cd pr-reviewer-service

docker-compose up