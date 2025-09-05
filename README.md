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

## Production Setup

For production deployment, create a `.env` file with the following required variables:

```bash
# Required for Production
ADMIN_PASSWORD=your_secure_admin_password_here
JWT_SECRET=your_jwt_secret_here_minimum_32_characters
ENCRYPTION_KEY=your_encryption_key_here_exactly_32_bytes

# Optional
BACKEND_PORT=6080
DB_PATH=/data/upm.db
LETSENCRYPT_EMAIL=your_email@example.com
```

**Important:** The application will exit immediately if these required variables are not set in production mode.

## Development Notes

- Database is automatically created
- Development mode uses default secrets (not secure for production)
- CORS enabled for development
- No `.env` files needed for development

## Disclaimer

**This software is for educational/development purposes only. Not recommended for production use without proper security review. See [DISCLAIMER.md](DISCLAIMER.md) for important security warnings.**

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
