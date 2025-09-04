# Undecided Proxy Manager (UPM)

A simple proxy management system with Go backend, Vue 3 frontend, and Swagger API documentation.

## Quick Start

```bash
# Development (hot reload)
./dev.sh up

# Production (optimized)
./prod.sh up
```

## Development Commands

```bash
# Start development environment (no build)
./dev.sh up

# Start with build (when you need to rebuild)
./dev.sh up-build

# Build containers only
./dev.sh build

# Stop development environment
./dev.sh down

# View logs
./dev.sh logs

# Restart development environment
./dev.sh restart
```

## Ports

| Environment | Frontend | Backend | Swagger |
|-------------|----------|---------|---------|
| **Development** | 6071 | 6081 | http://localhost:6081/swagger |
| **Production** | 6070 | 6080 | http://localhost:6080/swagger |

## Manual Development

```bash
# Backend only
cd backend && go run main.go

# Frontend only  
cd frontend && yarn dev
```

## Notes

- Database is automatically created
- All endpoints return mock data (database integration pending)
- CORS enabled for development
- No `.env` files needed

## Disclaimer

**This software is for educational/development purposes only. Not recommended for production use without proper security review. See [DISCLAIMER.md](DISCLAIMER.md) for important security warnings.**

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
