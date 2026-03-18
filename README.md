# topikkhusus-methodtracker

## 1. Project Overview
This project is a production-oriented REST API built with Go and Redis. It provides user management endpoints and a method tracking system that logs every incoming HTTP request into Redis.

## 2. Architecture
The codebase follows a layered architecture:

- Handler layer: HTTP parsing, validation, and response writing.
- Service layer: business rules and orchestration.
- Repository layer: Redis persistence implementation.
- Middleware layer: request ID, logger, tracker, recovery, and rate limiting.
- Infrastructure layer: config loading and Redis client initialization.

Dependency Injection is used end-to-end from `main.go`, so there is no global mutable state.

## 3. Folder Structure
```text
topikkhusus-methodtracker/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── models/
│   │   └── user.go
│   ├── repository/
│   │   └── user_repository.go
│   ├── services/
│   │   └── user_service.go
│   ├── handlers/
│   │   └── user_handler.go
│   ├── middleware/
│   │   └── middleware.go
│   └── tracker/
│       └── tracker.go
├── pkg/
│   └── redis/
│       └── redis.go
├── routes/
│   └── routes.go
├── tests/
│   └── user_service_test.go
├── .env.example
├── Makefile
├── go.mod
└── README.md
```

## 4. Redis Data Structure Explanation
This project uses Redis only:

- User data (Hash): `user:{id}`
  - Fields: `id`, `name`, `email`, `created_at`
- User index (Set): `users:index`
  - Members: all user IDs
- Method track logs (List): `method:tracks`
  - Each item is a JSON object
  - Insertion uses `LPUSH`
  - Retention uses `LTRIM 0 999` (max 1000 entries)

## 5. Method Tracking Explanation
Every request goes through middleware and is tracked with this shape:

```json
{
  "request_id": "uuid",
  "endpoint": "/users",
  "method": "POST",
  "status_code": 201,
  "execution_time_ms": 12,
  "timestamp": "2026-01-01T10:00:00Z"
}
```

Middleware execution order:

1. RequestID
2. Logger
3. Tracker
4. Recovery

Rate limiting middleware is also applied globally to protect endpoints.

## 6. Installation & Setup (Linux/Arch based)
1. Install Go and Redis:

```bash
sudo pacman -Syu
sudo pacman -S go redis
```

2. Start Redis service:

```bash
sudo systemctl enable redis
sudo systemctl start redis
```

3. Prepare environment file:

```bash
cp .env.example .env
```

4. Download dependencies:

```bash
go mod tidy
```

## 7. How to Run
Run server:

```bash
make run
```

Build binary:

```bash
make build
```

Run tests:

```bash
make test
```

Run lint check:

```bash
make lint
```

## 8. API Documentation with curl Examples
Base URL:

```text
http://localhost:8080
```

### Health Check
```bash
curl -i http://localhost:8080/health
```

### Create User
```bash
curl -i -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

### Get All Users
```bash
curl -i http://localhost:8080/users
```

### Get User by ID
```bash
curl -i http://localhost:8080/users/<user_id>
```

### Delete User by ID
```bash
curl -i -X DELETE http://localhost:8080/users/<user_id>
```

Response format:

- Success: `{"success": true, "data": ...}`
- Error: `{"success": false, "error": "message"}`

## 9. Troubleshooting
- `redis ping failed`:
  - Ensure Redis is running and `.env` host/port are correct.
- `address already in use`:
  - Change `APP_PORT` in `.env`.
- `too many requests`:
  - Increase `RATE_LIMIT_RPM` in `.env` for local testing.
- `unsupported media type`:
  - Ensure `Content-Type: application/json` is included for `POST /users`.

