# Skeleton Go API

This project is a Go + Fiber REST skeleton with clean layering (controllers, usecases, repositories) and structured logging. It now ships with a local MySQL database running in Docker plus JWT-based authentication.

## Getting Started

1. **Run dependencies via Docker**
   ```bash
   docker compose up -d mysql redis
   ```

2. **Apply database migrations**
   ```bash
   go run ./cmd/migrate
   ```
   This applies every SQL file under `migrations/`, creates the schema, and seeds the admin account.

3. **Run the API**
   ```bash
   go run ./...
   ```
   The app reads configuration from `resources/config.local.yaml` by default (APP_ENV=local) which already matches the Docker credentials/port (user `skeleton`, password `secret`, DB `skeleton_local`).

### Database access

The API uses GORM for all repositories while sharing the same `database/sql` connection pool. Flip `database.log_queries` in the config files if you want ORM trace logging in your environment.

### Authentication & refresh tokens

Login returns both an access token (used in the `Authorization` header) and a refresh token. Only one refresh token can be active per user—logging in rotates both tokens and invalidates the previous session. Exchange refresh tokens via `POST /api/v1/auth/refresh` to obtain a new pair without re-sending credentials.

Call `POST /api/v1/auth/logout` (with your access token) to revoke the refresh token and force a new login. Configure TTLs through the `auth` section in `resources/config.<env>.yaml` and point the refresh-token store at your Redis deployment via `redis.*` values.

### Third-party example endpoint

Set `external.jsonplaceholder_url` if you want to point the sample integration somewhere else. The proxy lives under `/api/v1` and requires the same auth token as the rest of the protected routes. Once you log in, hit it with:

```bash
curl http://localhost:8080/api/v1/placeholder/posts/1 \
  -H 'Authorization: Bearer <TOKEN>'
```

The handler proxies JSONPlaceholder and surfaces the response through the same response envelope as the rest of the app.

## Seeded Credentials

The Docker init script seeds a user you can use to obtain an auth token:

- Email: `admin@example.com`
- Password: `admin1234`

Authenticate:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"admin1234"}'
```
Use the returned JWT as a Bearer token (or `X-Auth-Token` header) for `/api/v1/users` endpoints.

## Localization

Response messages are localized by default in Indonesian (`locale: "id"`), with English translations already provided. Edit or add dictionaries under `resources/i18n/messages.<locale>.yaml` and update `app.locale` in the relevant `resources/config.<env>.yaml` file to change defaults or add more languages.

## Environment Notes

- Change `resources/config.<env>.yaml` to customize connection info, logging rotation, or JWT secrets.
- The compose volume `mysql_data` persists DB data between runs; remove it (`docker volume rm skeleton-go_mysql_data`) to reset.

## License

Released under the MIT License. See `LICENSE` for details.
