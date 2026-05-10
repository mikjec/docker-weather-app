# syntax=docker/dockerfile:1

#Etap 1
FROM golang:1.22-alpine AS builder

# Instalacja narzędzi do SSH i Gita
RUN apk add --no-cache git openssh-client
RUN mkdir -p -m 0700 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts

WORKDIR /app

# Pobieranie kodu przez SSH
RUN --mount=type=ssh git clone git@github.com:mikjec/docker-weather-app.git .

# Statyczna kompilacja pod wiele architektur
RUN CGO_ENABLED=0 go build -o weather-app main.go


#Etap 2

FROM alpine:3.19
LABEL org.opencontainers.image.authors="Mikolaj Jeczala"
WORKDIR /root/
COPY --from=builder /app/weather-app .
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./weather-app"]
