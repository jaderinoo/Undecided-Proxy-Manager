#!/bin/bash

# UPM Development Script
# Usage: ./dev.sh [up|down|logs|restart]

set -e

case "${1:-up}" in
  up)
    echo "Starting UPM in development mode..."
    docker-compose -f docker-compose.dev.yml up
    ;;
  up-build)
    echo "Starting UPM in development mode (with build)..."
    docker-compose -f docker-compose.dev.yml up --build
    ;;
  down)
    echo "Stopping UPM development environment..."
    docker-compose -f docker-compose.dev.yml down
    ;;
  logs)
    echo "Showing UPM development logs..."
    docker-compose -f docker-compose.dev.yml logs -f
    ;;
  restart)
    echo "Restarting UPM development environment..."
    docker-compose -f docker-compose.dev.yml down
    docker-compose -f docker-compose.dev.yml up
    ;;
  build)
    echo "Building UPM development environment..."
    docker-compose -f docker-compose.dev.yml build
    ;;
  *)
    echo "Usage: $0 [up|up-build|down|logs|restart|build]"
    echo ""
    echo "Commands:"
    echo "  up       - Start development environment (default, no build)"
    echo "  up-build - Start development environment with build"
    echo "  down     - Stop development environment"
    echo "  logs     - Show logs"
    echo "  restart  - Restart development environment"
    echo "  build    - Build development environment only"
    exit 1
    ;;
esac
