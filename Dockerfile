FROM golang:1.24.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Create required directories
RUN mkdir -p /app/simData/log_data
RUN mkdir -p /root/backend/initDB

# First check if the source directory exists
RUN ls -la /app/backend/connections/initDB || echo "Source directory not found"

# Then list the contents to see what files are available
RUN find /app/backend/connections -name "*.sql" -type f | sort

# Copy SQL files to the expected location with better error handling
RUN if [ -d "/app/backend/connections/initDB" ] && [ "$(ls -A /app/backend/connections/initDB)" ]; then \
    cp -v /app/backend/connections/initDB/*.sql /root/backend/initDB/ || exit 1; \
    else \
    echo "ERROR: Source directory empty or not found"; \
    exit 1; \
    fi

# Verify the files were copied correctly
RUN ls -la /root/backend/initDB/

ENV APP_ENV=prod \
    PORT=8080 \
    DOCKER_ENV=true \
    DB_HOST=postgres \
    DB_USER=postgres \
    DB_PASSWORD=Week7890 \
    DB_NAME=summervilledb \
    DB_PORT=5432 \
    DB_SSLMODE=disable \
    KAFKA_BROKER=kafka:29092 \
    REDIS_HOST=redis \
    REDIS_PORT=6379 \
    SQL_INIT_PATH=/root/backend/initDB

EXPOSE 8080

CMD ["go", "run", "."]