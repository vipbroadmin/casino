FROM golang:1.22-alpine AS build

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /players_service ./cmd

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=build /players_service /app/players_service
COPY migrations /app/migrations

EXPOSE 8088

CMD ["/app/players_service"]
