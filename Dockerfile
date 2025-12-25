# Build stage
FROM golang:1.25.5-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /taskflow ./cmd/api

# Run stage
FROM alpine:3.20
WORKDIR /app
COPY --from=build /taskflow /usr/local/bin/taskflow
EXPOSE 8080
ENTRYPOINT ["taskflow"]
