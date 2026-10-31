# Trading Platform

This is a fake trading platform, built specifically as a target for autotests — not a real broker. It is a Go monorepo: an API gateway plus three services that share one Postgres. Java autotests should hit the gateway; URL prefixes stay stable when more services appear.

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

## API

Java autotests should call the gateway at `http://localhost:8080`. Each service also serves the same paths on its debug port. Send `Content-Type: application/json` when the endpoint has a body. Protected routes need `Authorization: Bearer <accessToken>` from login.

### api-gateway

`http://localhost:8080`. These routes are handled by the gateway and are not proxied.

**GET** `http://localhost:8080/health`

No request body. Liveness of the gateway itself.

**GET** `http://localhost:8080/ready`

No request body. Calls `/health` on auth-service, user-service, and market-data-service.

**GET** `http://localhost:8080/openapi.yaml`

No request body.

**GET** `http://localhost:8080/swagger`

No request body.

### auth-service

Gateway prefix `/api/v1/auth`. Debug: `http://localhost:8082`.

**POST** `http://localhost:8080/api/v1/auth/register`

Public. The created role is always `user`. `displayName` is optional.

```json
{
  "email": "trader@example.com",
  "password": "secret123",
  "displayName": "Trader"
}
```

**POST** `http://localhost:8080/api/v1/auth/login`

Public. Returns `accessToken`, `tokenType` (`Bearer`), and `expiresIn` (seconds).

```json
{
  "email": "admin@example.com",
  "password": "admin123"
}
```

**GET** `http://localhost:8082/health`

No request body. Service liveness; the gateway does not proxy this path (it has its own `/health`).

### user-service

Gateway prefix `/api/v1/users`. Debug: `http://localhost:8083`. Every route except `/health` requires a JWT.

**GET** `http://localhost:8080/api/v1/users/me`

Any authenticated user. No request body.

**PATCH** `http://localhost:8080/api/v1/users/me`

Any authenticated user. `displayName` is required; an empty string clears it. At most 100 characters.

```json
{
  "displayName": "New Name"
}
```

**GET** `http://localhost:8080/api/v1/users?page=1&pageSize=20`

Admin. No request body. Query: `page` (default `1`), `pageSize` (default `20`, max `100`).

**DELETE** `http://localhost:8080/api/v1/users/{id}`

Admin. No request body. `{id}` is a UUID. Deleting your own id returns `409 CONFLICT`.

**GET** `http://localhost:8083/health`

No request body.

### market-data-service

Gateway prefix `/api/v1/instruments`. Debug: `http://localhost:8084`. Every route except `/health` requires a JWT. `assetClass` is one of `stock`, `crypto`, `forex`, `other`. `symbol` is 1–32 characters: letters, digits, `.` or `-` (stored uppercased).

**GET** `http://localhost:8080/api/v1/instruments?page=1&pageSize=20`

Any authenticated user. No request body. Query: `page` (default `1`), `pageSize` (default `20`, max `100`).

**GET** `http://localhost:8080/api/v1/instruments/{id}`

Any authenticated user. No request body. `{id}` is a UUID.

**POST** `http://localhost:8080/api/v1/instruments`

Admin.

```json
{
  "symbol": "MSFT",
  "name": "Microsoft Corporation",
  "assetClass": "stock"
}
```

**PUT** `http://localhost:8080/api/v1/instruments/{id}`

Admin. Replaces the instrument. Same body as create. `{id}` is a UUID.

```json
{
  "symbol": "MSFT",
  "name": "Microsoft Corp.",
  "assetClass": "stock"
}
```

**DELETE** `http://localhost:8080/api/v1/instruments/{id}`

Admin. No request body. `{id}` is a UUID.

**GET** `http://localhost:8084/health`

No request body.

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
