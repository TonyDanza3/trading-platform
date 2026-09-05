# Trading Platform

Go monorepo: API gateway plus three services that share one Postgres. Java autotests should hit the gateway; URL prefixes stay stable when more services appear.

## Run

```bash
docker compose -f deployments/docker-compose.yml up --build
```

- Gateway (public): http://localhost:8080
- auth-service (debug): http://localhost:8082
- user-service (debug): http://localhost:8083
- market-data-service (debug): http://localhost:8084
- Health: http://localhost:8080/health
- Ready: http://localhost:8080/ready
- Swagger UI: http://localhost:8080/swagger
- OpenAPI: http://localhost:8080/openapi.yaml

Seeded admin: `admin@example.com` / `admin123`  
Seeded instruments: `AAPL`, `BTC-USD`, `EURUSD`.

## Layout

```text
pkg/{web,authjwt,httpx,dbx}
services/api-gateway
services/auth-service
services/user-service
services/market-data-service
api/openapi.yaml
```

| Prefix | Service |
|---|---|
| `/api/v1/auth` | auth-service |
| `/api/v1/users` | user-service |
| `/api/v1/instruments` | market-data-service |

Gateway forwards `Authorization` and `X-Request-Id`. JWT is checked by the service that owns the route. Auth and user share the `users` table; market-data owns `instruments`.

Unknown prefixes (for example `/api/v1/orders`) return `404 NOT_FOUND`. A down upstream returns `502 BAD_GATEWAY`.

## Endpoints

| Method | Path | Who |
|---|---|---|
| POST | `/api/v1/auth/register` | public |
| POST | `/api/v1/auth/login` | public |
| GET/PATCH | `/api/v1/users/me` | any user |
| GET | `/api/v1/users` | admin |
| DELETE | `/api/v1/users/{id}` | admin (cannot delete self) |
| GET | `/api/v1/instruments` | any user |
| GET | `/api/v1/instruments/{id}` | any user |
| POST/PUT/DELETE | `/api/v1/instruments` | admin |

## Error body

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [{"field": "email", "message": "email is required"}]
  }
}
```

Codes: `VALIDATION_ERROR` (400), `UNAUTHORIZED` (401), `FORBIDDEN` (403), `NOT_FOUND` (404), `CONFLICT` (409), `BAD_GATEWAY` (502), `UNAVAILABLE` (503), `INTERNAL_ERROR` (500).

## Example

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"admin123"}'
```
