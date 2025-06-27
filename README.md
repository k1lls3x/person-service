# Person Service

This service enriches person data with age, gender and nationality using external APIs.

## Running

1. Copy `example.env` to `.env` and set the environment variables.
2. Apply migrations using `migrate` (https://github.com/golang-migrate/migrate):
   ```
   migrate -path migrations -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" up
   ```
3. Build and run the server:
   ```
   go run ./cmd/server
   ```

## Configuration

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` – database connection.
- `AGE_API_URL`, `GENDER_API_URL`, `NATIONALITY_API_URL` – endpoints of external services.
- `LOG_LEVEL` – logging level (`debug`, `info`, `warn`, `error`).
