# hhraiser

Automatically raises your HeadHunter resume at scheduled times. A free alternative to the paid [Продвижение.LITE](https://hh.ru/applicant/services/payment?from=landing&package=lite) service.

## Features

- Scheduled resume raising at configurable times of day
- Random jitter between raises to avoid anti-fraud detection
- Timezone-aware scheduling

## Configuration

Configuration is provided via environment variables or a `.env` file placed in the `/config` volume.

| Variable | Required | Default | Description |
|---|---|---|---|
| `TZ` | | `UTC` | Timezone in [IANA format](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) |
| `LOG_LEVEL` | | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `HH_PHONE` | ✓ | | HeadHunter account phone number |
| `HH_PASSWORD` | ✓ | | HeadHunter account password |
| `HH_RESUME_ID` | ✓ | | Resume ID to raise |
| `HH_RESUME_TITLE` | | | Resume display name used in notifications |
| `SCHEDULE_TIMES` | ✓ | | Comma-separated raise times in `HH:MM` format |
| `SCHEDULE_JITTER` | | `5m` | Maximum random delay before each raise |
| `HTTP_TIMEOUT` | | `10s` | HTTP requests timeout |

### .env file

As an alternative to environment variables, you can place a `.env` file in the `/config` directory:

```env
TZ=Europe/Moscow
HH_PHONE=+79991234567
HH_PASSWORD=your_password
HH_RESUME_ID=abc123def456
HH_RESUME_TITLE=Go Developer
SCHEDULE_TIMES=10:00,14:05,18:10
SCHEDULE_JITTER=5m
```

Environment variables take precedence over the `.env` file.

### Finding your Resume ID

Open your resume on [hh.ru](https://hh.ru). The ID is the alphanumeric string in the URL: `https://hh.ru/resume/`**`abc123def456`**

## Building from Source

```bash
git clone https://github.com/rycln/hhraiser
cd hhraiser
go build -o hhraiser ./cmd/hhraiser
```

## License

MIT

