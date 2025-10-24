# Changelog

## [1.1.1] - 2025-10-24

### Fixed
- CSV export filename for problems endpoint (was incorrectly using "requests.csv" instead of "problems.csv")

## [1.1.0] - 2025-10-24

### Added
- Render.com deployment support with blueprint configuration (render.yaml)
- Comprehensive deployment documentation (DEPLOYMENT.md)
- .dockerignore file for optimized Docker builds
- Deployment section in README with Render.com and Docker instructions

### Changed
- Updated README with detailed deployment and Docker sections
- Improved documentation for production deployment scenarios

## [1.0.0] - 2025-10-24

### Added
- Docker support with multi-stage build
- Docker Compose configuration for easy deployment
- Configurable database path via environment variable (DB_PATH)
- Health check endpoint with HEAD request support
- Complete API monitoring with Jikan proxy
- Request logging and problem detection
- Swagger API documentation

### Changed
- Database path now configurable via DB_PATH environment variable

### Fixed
- Health check endpoint now handles HEAD requests for Docker health checks
