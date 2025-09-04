# Undecided Proxy Manager (UPM)

A simple proxy management system with Go backend, Vue 3 frontend, and Swagger API documentation.

## 🚀 Super Easy Setup

```bash
# Development (hot reload)
./dev.sh up

# Production (optimized)
./prod.sh up
```

## 📍 Ports

| Environment | Frontend | Backend | Swagger |
|-------------|----------|---------|---------|
| **Development** | 6071 | 6081 | http://localhost:6081/swagger |
| **Production** | 80 | 6080 | http://localhost:6080/swagger |

## 🔧 What You Get

- **REST API** with full CRUD operations for proxies and users
- **Interactive Swagger docs** - test APIs directly in your browser
- **Hot reload** in development mode
- **No config files needed** - everything just works
- **Pi-hole friendly** - uses port 6080 to avoid conflicts

## 📚 API Examples

```bash
# Health check
curl http://localhost:6081/health

# Get all proxies
curl http://localhost:6081/api/v1/proxies

# Create a proxy
curl -X POST http://localhost:6081/api/v1/proxies \
  -H "Content-Type: application/json" \
  -d '{"name":"test","domain":"example.com","target_url":"http://localhost:6071","ssl_enabled":false}'
```

## 🛠️ Manual Development

```bash
# Backend only
cd backend && go run main.go

# Frontend only  
cd frontend && yarn dev
```

## 📝 Notes

- Database is automatically created
- All endpoints return mock data (database integration pending)
- CORS enabled for development
- No `.env` files needed

## ⚠️ Disclaimer

**This software is for educational/development purposes only. Not recommended for production use without proper security review. See [DISCLAIMER.md](DISCLAIMER.md) for important security warnings.**

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
