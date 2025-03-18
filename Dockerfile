#FROM states that we are using the golang image as the base image
FROM golang:1.24.1 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

COPY go.mod ./

#Builds dependencies
RUN go mod download

# Download all files except those in .dockerignore
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o app

# Use a smaller image for the final stage
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/
COPY --from=builder /app/app .
COPY --from=builder /app/web/dist ./web/dist
COPY --from=builder /app/web/src ./web/src

# Copy any config files needed
COPY --from=builder /app/.env .
COPY --from=builder /app/connections.json .

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./app"]