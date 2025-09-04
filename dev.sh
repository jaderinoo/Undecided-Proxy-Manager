#!/bin/bash

# UPM Development Script
# Usage: ./dev.sh [up|down|logs|restart]

set -e

case "${1:-up}" in
  up)
    echo "🚀 Starting UPM in development mode..."
    docker-compose -f docker-compose.dev.yml up --build
    ;;
  down)
    echo "🛑 Stopping UPM development environment..."
    docker-compose -f docker-compose.dev.yml down
    ;;
  logs)
    echo "📋 Showing UPM development logs..."
    docker-compose -f docker-compose.dev.yml logs -f
    ;;
  restart)
    echo "🔄 Restarting UPM development environment..."
    docker-compose -f docker-compose.dev.yml down
    docker-compose -f docker-compose.dev.yml up --build
    ;;
  *)
    echo "Usage: $0 [up|down|logs|restart]"
    echo ""
    echo "Commands:"
    echo "  up       - Start development environment (default)"
    echo "  down     - Stop development environment"
    echo "  logs     - Show logs"
    echo "  restart  - Restart development environment"
    exit 1
    ;;
esac
