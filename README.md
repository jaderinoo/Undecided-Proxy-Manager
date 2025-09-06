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
| **Production** | ${PROD_FRONTEND_PORT:-6070} | ${PROD_BACKEND_PORT:-6080} | http://localhost:${PROD_BACKEND_PORT:-6080}/swagger |

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
JWT_SECRET=your_jwt_secret_here_minimum_32_characters
ENCRYPTION_KEY=your_encryption_key_here_exactly_32_bytes

# Optional
ADMIN_PASSWORD=your_secure_admin_password_here  # If not set, will use default password
BACKEND_PORT=6080
DB_PATH=/data/upm.db
LETSENCRYPT_EMAIL=your_email@example.com
```

**Important:** The application will exit immediately if the required variables are not set in production mode.

### Admin User Management

The application maintains a 1:1 relationship between the `ADMIN_PASSWORD` environment variable and the admin user:

- **If `ADMIN_PASSWORD` is set:** Creates/updates admin user with the provided password
- **If `ADMIN_PASSWORD` is not set:** Removes admin user (disables admin access)
- **Password changes:** Automatically updates admin user password when `ADMIN_PASSWORD` changes
- **Admin credentials:** Username: `admin`, Email: `admin@upm.local`

**Behavior:**
- Setting `ADMIN_PASSWORD` for the first time → Creates admin user
- Changing `ADMIN_PASSWORD` → Updates admin user password
- Removing `ADMIN_PASSWORD` → Deletes admin user (disables admin access)
- Restarting with no `ADMIN_PASSWORD` → Admin access remains disabled

### Password Hashing

The application uses bcrypt for secure password hashing. To generate a hashed password for the `ADMIN_PASSWORD` variable:

```bash
# From the project root directory
cd backend
go run cmd/hash-password/main.go "your_secure_password_here"
```

This will output:
- The original password (for verification)
- The bcrypt hashed password
- A ready-to-use `.env` line with the hashed password

**Example:**
```bash
$ go run cmd/hash-password/main.go "mySecurePassword123"
Original password: mySecurePassword123
Hashed password: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy

Add this to your .env file:
ADMIN_PASSWORD=$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
```

**Security Notes:**
- Never use plain text passwords in production
- The bcrypt cost factor is set to the default (10) for good security/performance balance
- Each password generates a unique hash due to random salt generation
- Store the hashed password in your `.env` file, not the plain text version

## Development Notes

- Database is automatically created
- Development mode uses default secrets (not secure for production)
- CORS enabled for development
- No `.env` files needed for development

## Disclaimer

**This software is for educational/development purposes only. Not recommended for production use without proper security review. See [DISCLAIMER.md](DISCLAIMER.md) for important security warnings.**

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
