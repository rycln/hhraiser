# hhraiser

Автоматически поднимает резюме на HeadHunter в заданное время.

## Возможности

- Подъём резюме по расписанию в заданные моменты времени
- Случайная задержка перед каждым подъёмом для обхода антифрода
- Уведомления об успехе и ошибках через webhook
- Уведомления о запуске и остановке приложения
- Поддержка часовых поясов
- Готов к запуску в Docker

## Быстрый старт

```yaml
services:
  hhraiser:
    image: ghcr.io/rycln/hhraiser:latest
    container_name: hhraiser
    environment:
      - TZ=Europe/Moscow
      - HH_PHONE=+79991234567
      - HH_PASSWORD=your_password
      - HH_RESUME_ID=abc123def456
      - HH_RESUME_TITLE=Golang-разработчик
      - SCHEDULE_TIMES=10:00,13:00,18:00
      - SCHEDULE_JITTER=5m
      - WEBHOOK_URL=http://apprise:8000/notify
      - WEBHOOK_SECRET=your_webhook_secret
      - WEBHOOK_NOTIFY_ON_SUCCESS=true
      - LOG_LEVEL=info
    volumes:
      - ./config:/config
    restart: unless-stopped
```

## Конфигурация

Конфигурация задаётся через переменные окружения или файл `.env`.

| Переменная | Обязательная | По умолчанию | Описание |
|---|---|---|---|
| `TZ` | | `UTC` | Часовой пояс в формате [IANA](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) |
| `LOG_LEVEL` | | `info` | Уровень логирования: `debug`, `info`, `warn`, `error` |
| `HH_PHONE` | ✓ | | Номер телефона от аккаунта HeadHunter |
| `HH_PASSWORD` | ✓ | | Пароль от аккаунта HeadHunter |
| `HH_RESUME_ID` | ✓ | | ID резюме для подъёма |
| `HH_RESUME_TITLE` | | | Название резюме для уведомлений |
| `SCHEDULE_TIMES` | ✓ | | Время подъёма через запятую в формате `HH:MM` |
| `SCHEDULE_JITTER` | | `5m` | Максимальная случайная задержка перед подъёмом |
| `WEBHOOK_URL` | | | URL для отправки уведомлений |
| `WEBHOOK_SECRET` | | | Bearer-токен для авторизации webhook запросов |
| `WEBHOOK_NOTIFY_ON_SUCCESS` | | `true` | Отправлять уведомление при успешном подъёме |
| `HTTP_TIMEOUT` | | `10s` | Таймаут HTTP запросов |

### Файл .env

Вместо переменных окружения можно использовать файл `.env` в директории `/config`:

```env
TZ=Europe/Moscow
HH_PHONE=+79991234567
HH_PASSWORD=your_password
HH_RESUME_ID=abc123def456
HH_RESUME_TITLE=Golang-разработчик
SCHEDULE_TIMES=10:00,13:00,18:00
SCHEDULE_JITTER=5m
```

Переменные окружения имеют приоритет над файлом `.env`.

### Как найти ID резюме

Откройте резюме на [hh.ru](https://hh.ru). ID — это буквенно-цифровая строка в URL: `https://hh.ru/resume/`**`abc123def456`**

## Уведомления

Приложение отправляет POST-запрос на `WEBHOOK_URL` в следующих случаях:

- Запуск и остановка приложения
- Успешный подъём резюме (можно отключить через `WEBHOOK_NOTIFY_ON_SUCCESS=false`)
- Ошибка при подъёме резюме

Примеры payload:

```json
{ "event": "app_started", "timestamp": "2026-04-21T10:00:00Z" }
{ "event": "raise_success", "resume_title": "Golang-разработчик", "timestamp": "2026-04-21T10:00:05Z" }
{ "event": "raise_failure", "resume_title": "Golang-разработчик", "status_code": 403, "timestamp": "2026-04-21T10:00:05Z" }
{ "event": "app_stopped", "timestamp": "2026-04-21T10:00:10Z" }
```

## Сборка из исходников

```bash
git clone https://github.com/rycln/hhraiser
cd hhraiser
go build -o hhraiser ./cmd/hhraiser
```

## Лицензия

MIT