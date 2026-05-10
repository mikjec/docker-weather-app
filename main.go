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
