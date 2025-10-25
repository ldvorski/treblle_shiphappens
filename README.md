# Jikan API Request Monitor v1.1.1

A Go backend service that proxies requests to the Jikan API (MyAnimeList unofficial API), logs request metrics, and provides REST endpoints to view, sort, filter, and search logged requests. Automatically detects and tracks slow (>= 400ms) and failed responses as "Problem" objects.
This project can be used as a template for testing any API or multiple APIs with minor modifications.

## Features

- **API Request Logging**: Every proxied request to Jikan API is logged with method, path, response status, response time, and timestamp
- **Problem Detection**: Automatically identifies slow responses (>2000ms) and creates problem records
- **List & Table Views**: View logged requests and problems in both list and table formats
- **Sorting**: Sort by `created_at` or `response_time`
- **Filtering**: Filter by method, response status, response time range, creation date range
- **Search**: Search requests by path
- **Pagination**: Support for limit and offset parameters

## Tech Stack

- **Language**: Go 1.25.3
- **Web Framework**: Gin
- **Database**: SQLite3
- **External API**: Jikan API v4 (https://api.jikan.moe/v4)

## Database Schema

### api_requests
- `id`: INTEGER PRIMARY KEY AUTOINCREMENT
- `method`: TEXT NOT NULL
- `path`: TEXT NOT NULL
- `response_status`: INTEGER NOT NULL
- `response_time_ms`: INTEGER NOT NULL
- `created_at`: DATETIME DEFAULT CURRENT_TIMESTAMP

### problems
- `id`: INTEGER PRIMARY KEY AUTOINCREMENT
- `request_id`: INTEGER NOT NULL (FK to api_requests)
- `problem_type`: TEXT NOT NULL
- `description`: TEXT NOT NULL
- `threshold_ms`: INTEGER NOT NULL
- `created_at`: DATETIME DEFAULT CURRENT_TIMESTAMP

## Installation

1. Clone the repository:
```bash
git clone https://github.com/ldvorski/treblle_shiphappens.git
cd treblle_shiphappens
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Health Check
```bash
GET /health
```

### Jikan API Proxy
```bash
GET /api/jikan/*path
```
Proxies requests to Jikan API and logs metrics.

**Examples:**
```bash
# Get anime list
curl http://localhost:8080/api/jikan/anime

# Get specific anime
curl http://localhost:8080/api/jikan/anime/1

# Get anime characters
curl http://localhost:8080/api/jikan/anime/1/characters

# Get top anime
curl http://localhost:8080/api/jikan/top/anime
```

### View Logged Requests

#### List View
```bash
GET /api/requests
```

**Query Parameters:**
- `sort`: `created_at` | `response_time` (default: `created_at`)
- `method`: Filter by HTTP method (e.g., `GET`)
- `response`: Filter by response status code (e.g., `200`, `404`)
- `min_time`: Minimum response time in milliseconds
- `max_time`: Maximum response time in milliseconds
- `created_after`: Filter by creation date (format: `YYYY-MM-DD`)
- `created_before`: Filter by creation date (format: `YYYY-MM-DD`)
- `search`: Search in path (partial match)
- `limit`: Number of results (default: 100)
- `offset`: Pagination offset (default: 0)

**Examples:**
```bash
# Get all requests
curl http://localhost:8080/api/requests

# Get requests sorted by response time
curl http://localhost:8080/api/requests?sort=response_time

# Get slow requests (>1000ms)
curl "http://localhost:8080/api/requests?min_time=1000"

# Get requests with 404 status
curl "http://localhost:8080/api/requests?response=404"

# Search for anime character requests
curl "http://localhost:8080/api/requests?search=characters"

# Combine filters
curl "http://localhost:8080/api/requests?sort=response_time&min_time=200&method=GET&limit=20"
```

#### Table View
```bash
GET /api/requests/table
```
Returns the same data in a table-structured format with columns and rows.

**Example:**
```bash
curl http://localhost:8080/api/requests/table?limit=10
```

#### CSV Export
```bash
GET /api/requests/csv
```
Returns the same data in a CSV format

**Example:**
```bash
curl http://localhost:8080/api/requests/csv?limit=10
```

### View Problems (Slow Responses, Failed Requests)

#### List View
```bash
GET /api/problems
```

**Query Parameters:** (same as `/api/requests`)

**Examples:**
```bash
# Get all problems
curl http://localhost:8080/api/problems

# Get problems sorted by response time
curl http://localhost:8080/api/problems?sort=response_time

# Get problems created today
curl "http://localhost:8080/api/problems?created_after=2025-10-23"

# Search for specific endpoint problems
curl "http://localhost:8080/api/problems?search=anime/1"
```

#### Table View
```bash
GET /api/problems/table
```
Returns problems in a table-structured format.

**Example:**
```bash
curl http://localhost:8080/api/problems/table
```

#### CSV Export
```bash
GET /api/problems/csv
```
Returns the same data in a CSV format

**Example:**
```bash
curl http://localhost:8080/api/problems/csv
```

## Response Examples

### List View Response
```json
{
  "data": [
    {
      "id": 1,
      "method": "GET",
      "path": "/anime/1/characters",
      "response": 200,
      "response_time": 2345,
      "created_at": "2025-10-23T14:30:00Z"
    }
  ],
  "meta": {
    "limit": 100,
    "offset": 0
  }
}
```

### Table View Response
```json
{
  "columns": ["method", "response", "path", "response_time", "created_at"],
  "rows": [
    {
      "method": "GET",
      "response": 200,
      "path": "/anime/1/characters",
      "response_time": 2345,
      "created_at": "2025-10-23T14:30:00Z"
    }
  ],
  "meta": {
    "limit": 100,
    "offset": 0
  }
}
```

### Problems Response
```json
{
  "data": [
    {
      "id": 1,
      "request_id": 1,
      "problem_type": "slow_response",
      "description": "Response time (2345ms) exceeded threshold (2000ms)",
      "threshold_ms": 2000,
      "created_at": "2025-10-23T14:30:00Z",
      "method": "GET",
      "response": 200,
      "path": "/anime/1/characters",
      "response_time": 2345
    }
  ],
  "meta": {
    "limit": 100,
    "offset": 0
  }
}
```

## Project Structure

```
treblle_project/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── database/
│   │   └── db.go                # Database connection and migrations
│   ├── models/
│   │   ├── request.go           # APIRequest model
│   │   └── problem.go           # Problem model
│   ├── repository/
│   │   ├── request_repository.go # Request data access
│   │   └── problem_repository.go # Problem data access
│   ├── jikan/
│   │   └── client.go            # Jikan API client
│   └── handlers/
│       ├── request_handler.go   # Request viewing endpoints
│       ├── problem_handler.go   # Problem viewing endpoints
│       └── jikan_handler.go     # Jikan proxy endpoint
├── go.mod
├── .gitignore
└── README.md
```

## Deployment

### Deploy to Render.com

This application is configured for one-click deployment to Render.com.

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy)

**Manual Deployment:**

1. Fork or clone this repository to your GitHub account
2. Create a [Render.com](https://render.com) account
3. Click "New +" → "Blueprint"
4. Connect your GitHub repository
5. Render will automatically detect `render.yaml` and configure the service
6. Click "Apply" to deploy

**Environment Variables:**
- `GIN_MODE`: Set to `release` for production
- `DB_PATH`: Database file path (default: `/var/data/api_monitor.db`)
- `PORT`: Application port (default: `8080`)

**Database:**
A persistent disk is automatically created and mounted at `/var/data` for SQLite database storage.

**Accessing Your Deployment:**
Your API will be available at: `https://trebble-api-monitor.onrender.com`

### Deploy with Docker Compose (Local/VPS)

See the Docker section below for local deployment.

## Development

### Building
```bash
go build -o api_monitor cmd/server/main.go
./api_monitor
```

### Running Tests (if implemented)
```bash
go test ./...
```

## Docker Deployment

### Local Development with Docker

```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Docker Configuration

The application includes:
- **Multi-stage Dockerfile** for optimized image size
- **docker-compose.yml** for easy local deployment
- **Persistent volume** for SQLite database at `./data/`
- **Health checks** for container monitoring

**Environment Variables:**
- `DB_PATH`: Database file path (default: `./api_monitor.db`)
- `GIN_MODE`: Gin framework mode (`debug` or `release`)
- `TZ`: Timezone (default: `UTC`)

## Notes

- The database file `api_monitor.db` is created automatically in the project root
- Slow response threshold is set to 400ms (0.4 seconds)
- Default pagination limit is 100 records
- All timestamps are stored in UTC
- The Jikan API has rate limiting; be mindful when making requests

## License

MIT

