# Changelog

## [1.0.0] - 2025-10-24

### Added
- Docker support with multi-stage build
- Docker Compose configuration for easy deployment
- Configurable database path via environment variable
- Health check endpoint with HEAD request support
- Complete API monitoring with Jikan proxy
- Request logging and problem detection
- Swagger API documentation

### Changed
- Database path now configurable via DB_PATH environment variable

### Fixed
- Health check endpoint now handles HEAD requests for Docker health checks