English | [简体中文](README_CN.md)

# Zeus - Zephyr Backend Service

Zeus is a high-performance backend service for the Zephyr project, built on the Gin framework, providing weather data API services. The project supports multiple weather data sources, including QWeather and OpenMeteo, with complete caching mechanisms and TLS security support. It will later enable Zephyr to use the Zeus server you build to obtain data.

## Main Features

- **Multiple Data Source Support**: Integrated QWeather and OpenMeteo weather data sources
- **High Performance**: Built on Gin framework, supports high-concurrency requests
- **Secure and Reliable**: Supports TLS encrypted transmission
- **Smart Caching**: Redis caching mechanism to improve response speed
- **City Search**: Supports global city search and geolocation queries
- **Weather Alerts**: Real-time weather alert information push
- **Monitoring Ready**: Built-in health check endpoint

## Project Architecture

```
Zeus/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── api/             # API handlers
│   ├── config/          # Configuration management
│   ├── models/          # Data models
│   └── providers/       # External service providers
├── pkg/
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── Makefile             # Build automation
```

## Quick Start

### Requirements

- Go 1.25+
- Redis 8.0+
- TLS certificates (production environment)

### Installation Steps

1. **Clone the project**
   ```bash
   git clone https://github.com/LanceHuang245/Zeus
   cd Zeus
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Configure environment variables**
   ```bash
   cp .env.example .env
   # Edit the .env file and fill in your configuration information
   ```

4. **Start the service**
   ```bash
   make run
   ```

After the service starts, visit `https://localhost:3899/api/v1/healthcheck` to verify that the service is running properly.

## Configuration

### Environment Variables

| Variable Name | Description | Default Value |
|---------------|-------------|---------------|
| `REDIS_ADDR` | Redis address | `127.0.0.1:6379` |
| `REDIS_PASSWORD` | Redis password | Empty |
| `REDIS_DB` | Redis database | `0` |
| `CACHE_TTL_MINUTES` | Cache TTL (minutes) | `30` |
| `QWEATHER_PROJECT_ID` | QWeather project ID | - |
| `QWEATHER_KEY_ID` | QWeather Key ID | - |
| `QWEATHER_PRIVATE_KEY` | QWeather private key | - |
| `QWEATHER_URL` | QWeather API address | `https://devapi.qweather.com/v7` |
| `SERVER_PORT` | Service port | `:3899` |
| `ENABLE_TLS` | Enable TLS | `true` |
| `CERT_FILE` | TLS certificate path | `./cert/zephyr.crt` |
| `KEY_FILE` | TLS private key path | `./cert/zephyr.key` |

## Development Guide

### Build Project

```bash
# Local build
make build

# Build Linux version
make build-linux

# Build Windows version
make build-windows

# Build all platforms
make build-all
```

### Code Standards

```bash
# Format code
make fmt

# Clean build files
make clean
```

## Contributing

1. Fork the project
2. Create a feature branch
3. Commit changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

⭐ If this project helps you, please give it a Star to support us!