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
```

## 3. Instrukcja i dowody działania

### a. Budowanie obrazu
Użyto buildera `docker-container` dla obsługi wielu platform i cache registry.
```bash
docker buildx build --platform linux/amd64,linux/arm64 --ssh default=~/.ssh/id_ed25519 -t mikjecz/weather-app:v1 --cache-to type=registry,ref=mikjecz/weather-app-cache,mode=max --cache-from type=registry,ref=mikjecz/weather-app-cache --push .
```

```
1 --cache-to type=registry,ref=mikjecz/weather-app-cache,mode=max --cache-from type=registry,ref=mikjecz/weather-app-cache --push .
[+] Building 184.0s (32/32) FINISHED                                                                               docker-container:testbuilder
 => [internal] load build definition from Dockerfile                                                                                       0.0s
 => => transferring dockerfile: 846B                                                                                                       0.0s
 => resolve image config for docker-image://docker.io/docker/dockerfile:1                                                                  0.6s
 => CACHED docker-image://docker.io/docker/dockerfile:1@sha256:2780b5c3bab67f1f76c781860de469442999ed1a0d7992a5efdf2cffc0e3d769            0.0s
 => => resolve docker.io/docker/dockerfile:1@sha256:2780b5c3bab67f1f76c781860de469442999ed1a0d7992a5efdf2cffc0e3d769                       0.0s
 => [linux/amd64 internal] load metadata for docker.io/library/golang:1.22-alpine                                                          0.2s
 => [linux/amd64 internal] load metadata for docker.io/library/alpine:3.19                                                                 0.4s
 => [linux/arm64 internal] load metadata for docker.io/library/alpine:3.19                                                                 0.5s
 => [linux/arm64 internal] load metadata for docker.io/library/golang:1.22-alpine                                                          0.5s
 => [internal] load .dockerignore                                                                                                          0.0s
 => => transferring context: 2B                                                                                                            0.0s
 => ERROR importing cache manifest from mikjecz/weather-app-cache                                                                          0.6s
 => [linux/arm64 builder 1/6] FROM docker.io/library/golang:1.22-alpine@sha256:1699c10032ca2582ec89a24a1312d986a3f094aed3d5c1147b19880afe  0.1s
 => => resolve docker.io/library/golang:1.22-alpine@sha256:1699c10032ca2582ec89a24a1312d986a3f094aed3d5c1147b19880afe40e052                0.1s
 => [linux/amd64 builder 1/6] FROM docker.io/library/golang:1.22-alpine@sha256:1699c10032ca2582ec89a24a1312d986a3f094aed3d5c1147b19880afe  0.1s
 => => resolve docker.io/library/golang:1.22-alpine@sha256:1699c10032ca2582ec89a24a1312d986a3f094aed3d5c1147b19880afe40e052                0.1s
 => [linux/amd64 stage-1 1/3] FROM docker.io/library/alpine:3.19@sha256:6baf43584bcb78f2e5847d1de515f23499913ac9f12bdf834811a3145eb11ca1   0.1s
 => => resolve docker.io/library/alpine:3.19@sha256:6baf43584bcb78f2e5847d1de515f23499913ac9f12bdf834811a3145eb11ca1                       0.1s
 => [linux/arm64 stage-1 1/3] FROM docker.io/library/alpine:3.19@sha256:6baf43584bcb78f2e5847d1de515f23499913ac9f12bdf834811a3145eb11ca1   0.1s
 => => resolve docker.io/library/alpine:3.19@sha256:6baf43584bcb78f2e5847d1de515f23499913ac9f12bdf834811a3145eb11ca1                       0.1s
 => CACHED [linux/arm64 stage-1 2/3] WORKDIR /root/                                                                                        0.0s
 => CACHED [linux/arm64 builder 2/6] RUN apk add --no-cache git openssh-client                                                             0.0s
 => CACHED [linux/arm64 builder 3/6] RUN mkdir -p -m 0700 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts                           0.0s
 => CACHED [linux/arm64 builder 4/6] WORKDIR /app                                                                                          0.0s
 => CACHED [linux/arm64 builder 5/6] RUN --mount=type=ssh git clone git@github.com:mikjec/docker-weather-app.git .                         0.0s
 => CACHED [linux/amd64 stage-1 2/3] WORKDIR /root/                                                                                        0.0s
 => CACHED [linux/amd64 builder 2/6] RUN apk add --no-cache git openssh-client                                                             0.0s
 => CACHED [linux/amd64 builder 3/6] RUN mkdir -p -m 0700 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts                           0.0s
 => CACHED [linux/amd64 builder 4/6] WORKDIR /app                                                                                          0.0s
 => CACHED [linux/amd64 builder 5/6] RUN --mount=type=ssh git clone git@github.com:mikjec/docker-weather-app.git .                         0.0s
 => CACHED [linux/amd64 builder 6/6] RUN CGO_ENABLED=0 go build -o weather-app main.go                                                     0.0s
 => CACHED [linux/amd64 stage-1 3/3] COPY --from=builder /app/weather-app .                                                                0.0s
 => [linux/arm64 builder 6/6] RUN CGO_ENABLED=0 go build -o weather-app main.go                                                          151.3s
 => [linux/arm64 stage-1 3/3] COPY --from=builder /app/weather-app .                                                                       0.1s
 => exporting to image                                                                                                                    10.1s
 => => exporting layers                                                                                                                    0.8s
 => => exporting manifest sha256:58dee153c557fd7b425166f93d78b067868e9c03a768cbbd0b0814bac47328a0                                          0.0s
 => => exporting config sha256:2bc0086a825f0096525ad4d24a5d59e6809e698010d953ef680abe512c0fd526                                            0.0s
 => => exporting attestation manifest sha256:a3923083f3de0e08b6fa3cfe5e1d2bc6994c7ad5f1fdb808f18c3fba81c08a20                              0.0s
 => => exporting manifest sha256:d74b1533c262db8e9e170f2791a4aff41dc1b9e49d23352b63b58eff52152db9                                          0.0s
 => => exporting config sha256:783ecdc0418ccf4ae1f9b9324d795fce4c060c468f50485125673a98c39a7afd                                            0.0s
 => => exporting attestation manifest sha256:cee60c35766946ac8f323f0194442cdb8e1aa87ba5ead743ac896d0ee8d292e6                              0.1s
 => => exporting manifest list sha256:b1651daf34ac155538e606cdccd556ed15e553ece644afcf707f3ed95cbd6780                                     0.0s
 => => pushing layers                                                                                                                      4.5s
 => => pushing manifest for docker.io/mikjecz/weather-app:v1@sha256:b1651daf34ac155538e606cdccd556ed15e553ece644afcf707f3ed95cbd6780       4.5s
 => exporting cache to registry                                                                                                           29.4s
 => => preparing build cache for export                                                                                                    6.7s
 => => sending cache export                                                                                                               22.8s
 => => writing layer sha256:26530716c789b6455341a98d56b48aed116bf4af4a1b78b5827e8687f023ff76                                               1.9s
 => => writing layer sha256:1f3e46996e2966e4faa5846e56e76e3748b7315e2ded61476c24403d592134f0                                               2.7s
 => => writing layer sha256:1f486b9ab9ec9f853bb54ebd9d46fdc6460018619035d6886f2b89ede0cbb4c2                                               3.1s
 => => writing layer sha256:17a39c0ba978cc27001e9c56a480f98106e1ab74bd56eb302f9fd4cf758ea43f                                               2.6s
 => => writing layer sha256:2d88d4855e1f4a40e4e5c3ee18994f26fc968c58ff273829c3d3f31ca4aa63c6                                               1.9s
 => => writing layer sha256:4861bab1ea04dbb3dd5482b1705d41beefe250163e513588e8a7529ed76d351c                                               1.0s
 => => writing layer sha256:4b529a1a428d25a73b0af951fb36bcacd86e749d85ad0e2218abf97f46788167                                               1.4s
 => => writing layer sha256:4d75fd4b73869ed224045c010cdec78756eefb6752a5a8e4804294009eac11e9                                               0.9s
 => => writing layer sha256:4dd2b198310cf833fc366417834d87f73752ebfc0e0a09ee43602c8355713181                                               1.1s
 => => writing layer sha256:4f4fb700ef54461cfa02571ae0db9a0dc1e0cdb5577484a6d75e68dc38e8acc1                                               0.3s
 => => writing layer sha256:52f827f723504aa3325bb5a54247f0dc4b92bb72569525bc951532c4ef679bd4                                               2.0s
 => => writing layer sha256:5711127a7748d32f5a69380c27daf1382f2c6674ea7a60d2a3e338818590fea1                                               1.3s
 => => writing layer sha256:59607db92610efd514d91db4096c72a160bfa2e9374c25ae3df16e94f7b5971e                                               1.2s
 => => writing layer sha256:5f837c998576dcb54bc285997f33fcc2166dff6aa48fe3a374da92474efd5fe8                                               1.0s
 => => writing layer sha256:90fc70e12d60da9fe07466871c454610a4e5c1031087182e69b164f64aacd1c4                                              13.2s
 => => writing layer sha256:95eec11c4fc7f2dabd2c3ed156f4875cca606978551d3acb65fcd0116448e03c                                               1.1s
 => => writing layer sha256:afa154b433c7f72db064d19e1bcfa84ee196ad29120328f6bdb2c5fbd7b8eeac                                               5.9s
 => => writing layer sha256:bab3f3ebb351851333623db366a48c5b59ce0a2b87dc471f29ecfd659874d19a                                               4.5s
 => => writing layer sha256:e5b6e5c262c1cb6b4a4de570bd4772f8c346d92b0fa5b1057b74c25fb25a8d62                                               4.1s
 => => writing layer sha256:f779f57ff09fb5f7c3d6eaba6ea2e903ec6dd588b8493856e2b2354791761746                                               1.2s
 => => writing layer sha256:fa1868c9f11e67c6a569d83fd91d32a555c8f736e46d134152ae38157607d910                                               1.2s
 => => writing layer sha256:fa9c872e230520c682f3271480d4c7f844bc1ee320888994b34b3824fc11ec19                                               2.1s
 => => writing layer sha256:fb661ee02ef12d737bd2840abf2133d5d16f8a40935a47c83da7b514d755ab83                                               1.1s
 => => writing config sha256:61503e1a9c019121d2af5038dfe459fa135194c5f2a66e97993d6435dfac855e                                              1.3s
 => => writing cache image manifest sha256:9732586554b1f4ef0b34e292705162aaa3b85301704903966f51d19bd7811751                                2.1s
 => [auth] mikjecz/weather-app:pull,push token for registry-1.docker.io                                                                    0.0s
 => [auth] mikjecz/lab:pull mikjecz/weather-app:pull,push token for registry-1.docker.io                                                   0.0s
 => [auth] mikjecz/weather-app-cache:pull,push token for registry-1.docker.io    
```

Podczas pierwszego budowania w logach pojawił się komunikat ERROR importing cache manifest from mikjecz/weather-app-cache. Jest to zachowanie oczekiwane i można je zignorować. Błąd wynika z faktu, że przy pierwszym uruchomieniu polecenia, na koncie Docker Hub nie istniał jeszcze obraz weather-app-cache, z którego BuildKit mógłby pobrać dane. Przy każdym kolejnym budowaniu błąd ten już nie występuje, a w jego miejsce pojawiają się komunikaty potwierdzające pomyślne zaimportowanie i wykorzystanie cache (oznaczone słowem CACHED).

### b. Uruchomienie kontenera
```bash
docker run -d -p 8080:8080 --name weather-app mikjec/weather-app:v1
```

### c. Pobranie informacji z logów
```
docker logs weather-app
```

```
mikolaj@fedora:~$ docker logs weather-app 
2026/05/10 22:51:59 Uruchomiono: 2026-05-10T22:51:59Z
2026/05/10 22:51:59 Autor: Mikolaj Jeczala
2026/05/10 22:51:59 Aplikacja nasłuchuje na porcie TCP: 8080
```

### d. Liczba warstw i rozmiar obrazu
```
docker images
docker history mikjec/weather-app:v1
```
```
mikolaj@fedora:~$ docker images mikjecz/weather-app:v1                                                                                                                                    i Info →   U  In Use
IMAGE                    ID             DISK USAGE   CONTENT SIZE   EXTRA
mikjecz/weather-app:v1   b1651daf34ac       23.6MB         7.85MB    U   
mikolaj@fedora:~$ docker history mikjecz/weather-app:v1
IMAGE          CREATED        CREATED BY                                      SIZE      COMMENT
b1651daf34ac   13 hours ago   CMD ["./weather-app"]                           0B        buildkit.dockerfile.v0
<missing>      13 hours ago   HEALTHCHECK &{["CMD-SHELL" "wget -qO- http:/…   0B        buildkit.dockerfile.v0
<missing>      13 hours ago   EXPOSE [8080/tcp]                               0B        buildkit.dockerfile.v0
<missing>      13 hours ago   COPY /app/weather-app . # buildkit              7.7MB     buildkit.dockerfile.v0
<missing>      13 hours ago   WORKDIR /root/                                  4.1kB     buildkit.dockerfile.v0
<missing>      13 hours ago   LABEL org.opencontainers.image.authors=Mikol…   0B        buildkit.dockerfile.v0
<missing>      7 months ago   CMD ["/bin/sh"]                                 0B        buildkit.dockerfile.v0
<missing>      7 months ago   ADD alpine-minirootfs-3.19.9-x86_64.tar.gz /…   8.08MB    buildkit.dockerfile.v0
mikolaj@fedora:~$ 

```

## 4. Zadanie dodatkowe nr. 3

Obraz został stworzony zgodnie z wymaganiami zadania dodatkowego nr.3. W procesie budowy wykorzystano builder na sterowniku `docker container` oraz zastosowano odpowiedni cache.


```
mikolaj@fedora:~$ docker buildx imagetools inspect mikjecz/weather-app:v1
Name:      docker.io/mikjecz/weather-app:v1
MediaType: application/vnd.oci.image.index.v1+json
Digest:    sha256:b1651daf34ac155538e606cdccd556ed15e553ece644afcf707f3ed95cbd6780
           
Manifests: 
  Name:        docker.io/mikjecz/weather-app:v1@sha256:58dee153c557fd7b425166f93d78b067868e9c03a768cbbd0b0814bac47328a0
  MediaType:   application/vnd.oci.image.manifest.v1+json
  Platform:    linux/amd64
               
  Name:        docker.io/mikjecz/weather-app:v1@sha256:d74b1533c262db8e9e170f2791a4aff41dc1b9e49d23352b63b58eff52152db9
  MediaType:   application/vnd.oci.image.manifest.v1+json
  Platform:    linux/arm64
               
```
