#!/bin/bash

# UPM Production Script
# Usage: ./prod.sh [up|down|logs|restart]

set -e

case "${1:-up}" in
  up)
    echo "🚀 Starting UPM in production mode..."
    docker-compose up --build -d
    ;;
  down)
    echo "🛑 Stopping UPM production environment..."
    docker-compose down
    ;;
  logs)
    echo "📋 Showing UPM production logs..."
    docker-compose logs -f
    ;;
  restart)
    echo "🔄 Restarting UPM production environment..."
    docker-compose down
    docker-compose up --build -d
    ;;
  *)
    echo "Usage: $0 [up|down|logs|restart]"
    echo ""
    echo "Commands:"
    echo "  up       - Start production environment (default)"
    echo "  down     - Stop production environment"
    echo "  logs     - Show logs"
    echo "  restart  - Restart production environment"
    exit 1
    ;;
esac
