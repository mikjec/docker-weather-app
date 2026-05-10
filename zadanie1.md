# Sprawozdanie: Programowanie Aplikacji w Chmurze Obliczeniowej - Zadanie 1 część obowiązkowa + nieobowiązkowa punkt 3

**Autor:** Mikolaj Jeczala

---

## 1. Opis aplikacji
Aplikacja została napisana w języku **Go**. Realizuje ona następujące funkcje:
- Po uruchomieniu wypisuje w logach: datę uruchomienia, imię i nazwisko autora oraz port.
- Udostępnia interfejs webowy (port 8080) z formularzem do wpisania miasta.
- Pobiera dane o pogodzie z zewnętrznego API wttr.in i wyświetla je na tej samej stronie.

### Kod źródłowy main.go

```
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const author = "Mikolaj Jeczala"
const port = "8080"

func main() {

	//Wyświetlanie logów startowych
	log.Printf("Uruchomiono: %s\n", time.Now().Format(time.RFC3339))
	log.Printf("Autor: %s\n", author)
	log.Printf("Aplikacja nasłuchuje na porcie TCP: %s\n", port)

	http.HandleFunc("/", pageHandler)
	http.HandleFunc("/health", healthHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	var weatherResult string

	// Jeśli podano lokalizację, pobierz dane z API
	if location != "" {
		resp, err := http.Get("https://wttr.in/" + location + "?format=3")
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			weatherResult = fmt.Sprintf("<div style='margin-top:20px; padding:10px; border:1px solid #ccc;'><strong>Wynik dla %s:</strong> %s</div>", location, string(body))
			resp.Body.Close()
		} else {
			weatherResult = "<p style='color:red;'>Błąd pobierania pogody.</p>"
		}
	}

	// Wyświetlany szablon html
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="pl">
	<head><meta charset="UTF-8"><title>Pogodynka Cloud</title></head>
	<body style="font-family: sans-serif; max-width: 500px; margin: 40px auto; text-align: center;">
	<h2>Aplikacja Pogodowa</h2>
	<p>Autor: %s</p>
	<form action="/" method="get">
	<input type="text" name="location" placeholder="Wpisz miasto..." required style="padding: 8px;">
	<input type="submit" value="Sprawdź" style="padding: 8px;">
	</form>
	%s
	</body>
	</html>
	`, author, weatherResult)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

```


## 2. Plik Dockerfile
Dockerfile wykorzystuje **multi-stage build** oraz rozszerzony frontend **BuildKit**.

```dockerfile
# syntax=docker/dockerfile:1

#Etap 1
FROM golang:1.22-alpine AS builder

# Instalacja narzędzi do SSH i Gita
RUN apk add --no-cache git openssh-client
RUN mkdir -p -m 0700 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts

WORKDIR /app

# Pobieranie kodu przez SSH
RUN --mount=type=ssh git clone git@github.com:mikjecz/docker-weather-app.git .

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
```

## 3. Instrukcja i dowody działania

### a. Budowanie obrazu
Użyto buildera `docker-container` dla obsługi wielu platform i cache registry.
```bash
docker buildx build --platform linux/amd64,linux/arm64 --ssh default=$HOME/.ssh/id_rsa -t [TWÓJ_LOGIN]/pogoda:v1 --cache-to type=registry,ref=[TWÓJ_LOGIN]/pogoda-cache,mode=max --cache-from type=registry,ref=[TWÓJ_LOGIN]/pogoda-cache --push .
```

### b. Uruchomienie kontenera
Użyto buildera `docker-container` dla obsługi wielu platform i cache registry.
```bash
docker run -d -p 8080:8080 --name weather-app [TWÓJ_LOGIN]/pogoda:v1
```

### c. Pobranie informacji z logów
```docker logs weather-app
```

### d. Liczba warstw i rozmiar obrazu
```docker images oraz docker history [TWÓJ_LOGIN]/pogoda:v1
```
