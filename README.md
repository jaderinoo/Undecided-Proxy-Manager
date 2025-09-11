# Undecided Proxy Manager (UPM)

A simple proxy management system with Go backend, Vue 3 frontend, and Swagger API documentation.

This Project is a work in progress. Use it at your own risk.

## Mission Goals

I wanted to create a user friendly proxy management system that is easy to use and maintain. It uses a simple GO REST API and a Vue 3 frontend. The hope is to make self-hosted proxy management easy for everyone.

## Features

- **Proxy Management**: Create and manage reverse proxies with custom domains
- **SSL/TLS Certificates**: Automatic Let's Encrypt certificate generation and renewal
- **Dynamic DNS**: Support for Namecheap Dynamic DNS with automatic IP updates
- **Nginx Integration**: Automatic nginx configuration generation and reloading
- **Docker Support**: Full containerization with docker-compose
- **Modern UI**: Clean Vue 3 frontend with Vuetify components
- **REST API**: Complete Swagger-documented API
- **Security**: JWT authentication, password hashing, and encrypted storage

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
ADMIN_PASSWORD=your_secure_admin_password_here  # Plain text password (auto-hashed)
BACKEND_PORT=6080
DB_PATH=/data/upm.db
LETSENCRYPT_EMAIL=your_email@example.com
PUBLIC_IP_SERVICE=https://api.ipify.org
NGINX_CONFIG_PATH=/etc/nginx/sites-available
NGINX_RELOAD_CMD=docker exec undecided-proxy-manager-nginx-1 nginx -s reload
NGINX_CONTAINER_NAME=undecided-proxy-manager-nginx-1
```

**Important:** The application will exit immediately if the required variables are not set in production mode.

### Environment Configuration

The easiest way to set up your environment is through the web interface:

1. Start the application in development mode: `./dev.sh up`
2. Navigate to the Settings page
3. Click "Generate .env Template" to copy a complete configuration
4. Paste the template into your `.env` file and customize as needed

Alternatively, you can generate encryption keys using the built-in tools:

```bash
# Generate a secure encryption key
cd backend && go run cmd/generate-encryption-key/main.go

# Generate a JWT secret (32+ characters)
openssl rand -base64 32
```

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

### Password Security

The application automatically handles password security:

- **Plain text in .env**: Use plain text passwords in your `.env` file (e.g., `ADMIN_PASSWORD=mySecurePassword123`)
- **Automatic hashing**: The app automatically hashes passwords using bcrypt before storing in the database
- **Secure storage**: Only hashed passwords are stored in the database, never plain text
- **Password updates**: Changing `ADMIN_PASSWORD` in `.env` automatically updates the stored hash

**Security Features:**
- bcrypt hashing with default cost factor (10)
- Random salt generation for each password
- Secure password verification during login

## Development Notes

- Database is automatically created
- Development mode uses default secrets (not secure for production)
- CORS enabled for development
- No `.env` files needed for development

## Disclaimer

**This software is for educational/development purposes only. Not recommended for production use without proper security review. See [DISCLAIMER.md](DISCLAIMER.md) for important security warnings.**

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
