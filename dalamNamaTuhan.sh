#!/bin/bash
# Setup & run ChristAPI with Docker
# Usage: ./dalamNamaTuhan.sh

set -e

echo "🙏 Bismillah... Starting ChristAPI setup..."
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 1. Check if Docker is running
echo "🐳 Checking Docker..."
if ! docker ps &> /dev/null; then
    echo -e "${RED}❌ Docker is not running. Please start Docker Desktop and try again.${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Docker is running${NC}"
echo ""

# 2. Check if .env exists, if not copy from .env.example
echo "⚙️  Checking environment variables..."
if [ ! -f .env ]; then
    if [ ! -f .env.example ]; then
        echo -e "${RED}❌ .env.example not found${NC}"
        exit 1
    fi
    echo "📋 .env not found, copying from .env.example..."
    cp .env.example .env
    echo -e "${GREEN}✅ .env created${NC}"
else
    echo -e "${GREEN}✅ .env already exists${NC}"
fi
echo ""

# 3. Build Docker image
echo "🔨 Building Docker image..."
docker compose build --no-cache
echo -e "${GREEN}✅ Build complete${NC}"
echo ""

# 4. Start services
echo "🚀 Starting services (postgres, api)..."
docker compose down 2>/dev/null || true
docker compose up -d
echo -e "${GREEN}✅ Services started${NC}"
echo ""

# 5. Wait for postgres to be healthy
echo "⏳ Waiting for PostgreSQL to be healthy..."
max_attempts=30
attempt=1
while [ $attempt -le $max_attempts ]; do
    if docker compose exec -T postgre-chrisapi pg_isready -U christ_user &> /dev/null; then
        echo -e "${GREEN}✅ PostgreSQL is healthy${NC}"
        break
    fi
    echo "  Attempt $attempt/$max_attempts..."
    sleep 1
    ((attempt++))
done
if [ $attempt -gt $max_attempts ]; then
    echo -e "${RED}❌ PostgreSQL failed to become healthy${NC}"
    exit 1
fi
echo ""

# 6. Run migrations
echo "🔄 Running database migrations..."
docker compose run --rm migrate -path=/migrations -database "postgres://christ_user:christ_password@postgre-chrisapi:5432/christ_db?sslmode=disable" up
echo -e "${GREEN}✅ Migrations complete${NC}"
echo ""

# 7. Show status
echo "📊 Service status:"
docker compose ps
echo ""

# 8. Show access info
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}🎉 ChristAPI is ready!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "📍 API Server:"
echo "   http://localhost:3001"
echo ""
echo "🗄️  Database:"
echo "   Host: localhost"
echo "   Port: 5433"
echo "   Database: christ_db"
echo "   User: christ_user"
echo "   Password: christ_password"
echo ""
echo "📚 Useful commands:"
echo "   docker compose logs -f                 # View logs"
echo "   docker compose exec golang-christapi sh # Access API container"
echo "   docker compose down                    # Stop services"
echo ""
echo "DBeaver connection string:"
echo "   postgres://christ_user:christ_password@localhost:5433/christ_db"
echo ""
