FROM golang:1.24.1

WORKDIR /app

COPY . .

RUN go mod download

# Copy configs
COPY .env .
COPY connections.json .

ENV APP_ENV=prod

EXPOSE 8080

CMD ["go", "run", "."]
