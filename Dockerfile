FROM golang:1.24.1

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Create necessary directories
RUN mkdir -p /app/simData/log_data

# Environment variables with sensible defaults
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
    REDIS_PORT=6379

EXPOSE 8080

# Command to run the application
CMD ["go", "run", "."]