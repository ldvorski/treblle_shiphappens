# API Documentation

This project uses [swaggo/swag](https://github.com/swaggo/swag) to generate OpenAPI 3.0 documentation from code annotations.

## Installation

Install the swag CLI tool:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## Generating Documentation

From the project root directory, run:

```bash
cd treblle_project
swag init -g cmd/server/main.go -o ./docs
```

This will generate:
- `docs/swagger.json` - OpenAPI 3.0 specification in JSON format
- `docs/swagger.yaml` - OpenAPI 3.0 specification in YAML format  
- `docs/docs.go` - Go code for embedding the spec

## Viewing Documentation

### Option 1: Swagger UI (Recommended)

Add these packages to your `go.mod`:

```bash
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
```

Then add to `cmd/server/main.go`:

```go
import (
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    _ "treblle_project/docs"  // Import generated docs
)

// After setting up routes, add:
r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

Access the interactive documentation at: http://localhost:8080/swagger/index.html

### Option 2: Swagger Editor

1. Go to https://editor.swagger.io/
2. Upload `docs/swagger.yaml` or paste its contents
3. View and interact with the API documentation

### Option 3: Redoc

For a clean, single-page documentation:

```bash
npx @redocly/cli build-docs docs/swagger.yaml -o api-docs.html
```

## API Endpoints

### Requests
- `GET /api/requests` - List API requests with filtering
- `GET /api/requests/table` - Get requests in table format
- `GET /api/requests/csv` - Download requests as CSV

### Problems
- `GET /api/problems` - List detected problems with filtering
- `GET /api/problems/table` - Get problems in table format
- `GET /api/problems/csv` - Download problems as CSV

### Jikan Proxy
- `GET /api/jikan/*path` - Proxy requests to Jikan API with monitoring

### Health
- `GET /health` - Health check endpoint

## Query Parameters

All list endpoints support these filters:

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `method` | string | HTTP method filter | `GET`, `POST` |
| `response` | int | Response status code | `200`, `404` |
| `min_time` | int | Min response time (ms) | `100` |
| `max_time` | int | Max response time (ms) | `1000` |
| `created_after` | string | Filter by date | `2024-01-15` |
| `created_before` | string | Filter by date | `2024-01-20` |
| `search` | string | Search in path | `anime` |
| `sort` | string | Sort field | `response_time`, `created_at` |
| `limit` | int | Max results (default: 100) | `50` |
| `offset` | int | Skip results (default: 0) | `10` |

## Example Requests

### List all 404 errors
```bash
curl "http://localhost:8080/api/problems?response=404"
```

### Get slow requests (>500ms)
```bash
curl "http://localhost:8080/api/requests?min_time=500&sort=response_time"
```

### Proxy to Jikan API
```bash
curl "http://localhost:8080/api/jikan/anime/1"
```

### Export problems as CSV
```bash
curl "http://localhost:8080/api/problems/csv" --output problems.csv
```

## Problem Types

The API automatically detects and logs these problems:

| Type | Status Code | Description |
|------|-------------|-------------|
| `not_found` | 404 | Resource not found |
| `forbidden` | 403 | Access forbidden |
| `bad_request` | 400 | Invalid request |
| `im_a_teapot` | 418 | Server is a teapot (RFC 2324) |
| `slow_response` | Any | Response time >= 400ms |

## Response Format

### List Endpoints
```json
{
  "data": [...],
  "meta": {
    "count": 10,
    "limit": 100,
    "offset": 0
  }
}
```

### Table Endpoints
```json
{
  "columns": ["field1", "field2", ...],
  "rows": [[value1, value2, ...], ...],
  "meta": {
    "count": 10,
    "limit": 100,
    "offset": 0
  }
}
```

### Error Response
```json
{
  "error": "Error message"
}
```

## Updating Documentation

When you modify API endpoints:

1. Update the swag annotations in your handler functions
2. Regenerate the documentation: `swag init -g cmd/server/main.go -o ./docs`
3. Restart the server to see changes in Swagger UI

## Notes

- The `@Router` paths in annotations should match your actual Gin routes
- Query parameters are defined with `@Param` annotations
- Response schemas use `@Success` and `@Failure` annotations
- Models are automatically documented from struct tags

