# FotoBudka Backend

Backend aplikacji FotoBudka napisany w Go. Obsługuje logowanie fotografa, sesje zdjęciowe, dostęp klienta kodem/linkiem, upload zdjęć przez presigned URL, asynchroniczne generowanie miniaturek/proofów z watermarkiem, wybór zdjęć przez klienta, płatność manualną, upload finali, generowanie ZIP oraz publiczne portfolio fotografa.

## Spis treści

- [Opis projektu](#opis-projektu)
- [Główne funkcje](#główne-funkcje)
- [Architektura](#architektura)
- [Stack](#stack)
- [Struktura projektu](#struktura-projektu)
- [Wymagania](#wymagania)
- [Konfiguracja środowiska](#konfiguracja-środowiska)
- [Uruchomienie lokalne](#uruchomienie-lokalne)
- [Migracje bazy danych](#migracje-bazy-danych)
- [Uruchamianie testów i build](#uruchamianie-testów-i-build)
- [Najważniejsze endpointy](#najważniejsze-endpointy)
- [Statusy sesji](#statusy-sesji)
- [Bezpieczeństwo](#bezpieczeństwo)
- [Znane follow-upy](#znane-follow-upy)

## Opis projektu

FotoBudka usprawnia proces realizacji sesji zdjęciowych. Fotograf tworzy sesję, wgrywa zdjęcia, system generuje proofy ze znakiem wodnym, klient wybiera zdjęcia i dodaje notatki, a fotograf dostarcza finalne zdjęcia w paczce ZIP.

Backendowe MVP obejmuje pełny flow od logowania fotografa do pobrania ZIP przez klienta. Projekt zakłada frontend w React/TypeScript/Tailwind, backend w Go, PostgreSQL, object storage S3/MinIO/R2 oraz osobny worker do zadań asynchronicznych.

## Główne funkcje

- Logowanie fotografa przez Google OAuth.
- JWT w HttpOnly cookie, bez localStorage.
- Profil fotografa z publicznym `username`, bio i social links.
- CRUD sesji fotografa.
- Dostęp klienta przez kod albo link.
- Regeneracja kodu/linku z unieważnieniem poprzedniego dostępu.
- CAPTCHA po błędnych próbach wpisania kodu.
- Presigned upload zdjęć source bez przechodzenia plików przez backend.
- Worker generujący thumb/proof z watermarkiem.
- Klientowy wybór zdjęć i notatki per zdjęcie.
- Submit wyboru i przejście do `waiting_for_payment`.
- Płatność manualna oznaczana przez fotografa.
- Upload finalnych zdjęć pod wybrane zdjęcia.
- Generowanie wersjonowanych paczek ZIP.
- Pobieranie najnowszej paczki ZIP przez klienta.
- Publiczne portfolio fotografa i galerie bez watermarków.
- Featured public galleries na stronie głównej.

## Architektura

Projekt ma dwa główne procesy:

```txt
cmd/api     - HTTP API
cmd/worker  - worker zadań asynchronicznych
```

API odpowiada za logikę biznesową, autoryzację, presigned URL-e, handlery REST i zapis metadanych w PostgreSQL.

Worker odpowiada za ciężkie operacje:

- generowanie miniaturek,
- generowanie proofów z watermarkiem,
- generowanie ZIP z finalnymi zdjęciami.

Pliki zdjęć i ZIP-y nie są przechowywane w bazie. Baza trzyma tylko metadane i object keys, a obiekty są w S3/MinIO/R2.

## Stack

- Go
- `net/http`
- PostgreSQL
- `pgx/v5`, `pgxpool`
- `sqlc`
- `golang-migrate`
- Redis
- S3-compatible object storage:
  - dev: MinIO
  - prod: Cloudflare R2
- Google OAuth
- reCAPTCHA
- Docker Compose

## Struktura projektu

Przykładowe najważniejsze katalogi:

```txt
cmd/
  api/
  worker/

internal/
  app/
    api/
    worker/
  auth/
  config/
  galleries/
  guard/
  http/
    handler/
    middleware/
    router/
  jobsworker/
  oauth/
  payments/
  platform/
    captcha/
    db/
    logger/
    redis/
    storage/
  repository/
  selections/
  sessionaccess/
  sessionphotos/

scripts/
  migrate-up.sh
  migrate-version.sh
  test.sh
  build.sh
```

Konwencja projektu:

- handlery są dzielone domenami,
- routery są dzielone domenami,
- repozytoria są dzielone domenami,
- use-case/service istnieje tam, gdzie logika wykracza poza pojedyncze query,
- protected endpointy fotografa używają `RequireAuth`,
- klientowe endpointy po wejściu kodem/linkiem używają klientowego cookie i middleware dostępu klienta.

## Wymagania

Lokalnie:

- Docker
- Docker Compose
- Go
- `sqlc`, jeśli generujesz kod lokalnie
- dostęp do Google OAuth credentials
- reCAPTCHA keys

Na produkcji:

- VPS albo inny host dla API i workera
- PostgreSQL
- Redis
- Cloudflare R2 albo inny S3-compatible storage
- reverse proxy z HTTPS, np. nginx
- domena API, np. `https://fotobudka-api.mmozoluk.com`

## Konfiguracja środowiska

Plik `.env` jest wymagany dla Compose i aplikacji.

Przykład produkcyjnego fragmentu:

```env
APP_NAME=FotoBudka
APP_ENV=prod

API_PORT=8020
API_ADDR=:8080

BASE_URL=https://fotobudka-api.mmozoluk.com
FRONTEND_ORIGIN=https://fotobudka.mmozoluk.com

POSTGRES_DB=fotobudka
POSTGRES_USER=fotobudka
POSTGRES_PASSWORD=<strong-password>
DB_URL=postgres://fotobudka:<strong-password>@postgres:5432/fotobudka?sslmode=disable

REDIS_URL=redis://redis:6379/0

S3_ENDPOINT=<account-id>.r2.cloudflarestorage.com
S3_ENDPOINT_PUBLIC=<account-id>.r2.cloudflarestorage.com
S3_BUCKET=fotobudka-prod
S3_ACCESS_KEY_ID=<r2-access-key-id>
S3_SECRET_ACCESS_KEY=<r2-secret-access-key>
S3_REGION=auto
S3_USE_PATH_STYLE=true
S3_USE_SSL=true

GOOGLE_OAUTH_CLIENT_ID=<google-client-id>
GOOGLE_OAUTH_CLIENT_SECRET=<google-client-secret>
GOOGLE_OAUTH_REDIRECT_URL=https://fotobudka-api.mmozoluk.com/api/auth/google/callback

JWT_SECRET=<minimum-32-random-chars>
JWT_ISSUER=fotobudka
JWT_AUDIENCE=fotobudka-photographer
JWT_CLIENT_AUDIENCE=fotobudka-client
JWT_TTL_HOURS=720

COOKIE_NAME=fotobudka_auth
COOKIE_CLIENT_NAME=fotobudka_client
COOKIE_DOMAIN=
COOKIE_SECURE=true

RECAPTCHA_SECRET_KEY=<recaptcha-secret-key>
CODE_LOGIN_ATTEMPTS_TTL=10m
CODE_LOGIN_CAPTCHA_THRESHOLD=2

SIEM_LOG_DIR=/logs
```

W produkcji `COOKIE_SECURE=true` jest obowiązkowe. `S3_ENDPOINT` nie powinien zawierać schematu `https://`, jeśli config aplikacji oczekuje samego hosta i buduje URL na podstawie `S3_USE_SSL=true`.

## Uruchomienie lokalne

Typowy flow dev:

```bash
./scripts/dev-up.sh
```

Jeśli używasz ręcznie Compose:

```bash
mkdir -p logs
LOCAL_UID="$(id -u)" LOCAL_GID="$(id -g)" docker compose \
  -f compose.yaml \
  -f compose.dev.yaml \
  up -d --build
```

Logi:

```bash
docker compose -f compose.yaml -f compose.dev.yaml logs -f api
docker compose -f compose.yaml -f compose.dev.yaml logs -f worker
```

Healthcheck:

```bash
curl -i http://localhost:8080/healthz
```

## Migracje bazy danych

Migracje są uruchamiane przez `golang-migrate`.

```bash
./scripts/migrate-up.sh
```

Sprawdzenie wersji:

```bash
./scripts/migrate-version.sh
```

Na produkcji migracje najlepiej uruchamiać po starcie `postgres` i przed startem/restartem `api` oraz `worker`.

Przykładowa bezpieczna kolejność pierwszego uruchomienia:

```bash
LOCAL_UID="$(id -u)" LOCAL_GID="$(id -g)" docker compose \
  -p fotobudka \
  -f compose.yaml \
  build api worker

LOCAL_UID="$(id -u)" LOCAL_GID="$(id -g)" docker compose \
  -p fotobudka \
  -f compose.yaml \
  up -d postgres redis

./scripts/migrate-up.sh

LOCAL_UID="$(id -u)" LOCAL_GID="$(id -g)" docker compose \
  -p fotobudka \
  -f compose.yaml \
  up -d api worker
```

## Uruchamianie testów i build

Testy:

```bash
go test ./...
```

Jeżeli używasz projektowego skryptu:

```bash
./scripts/test.sh
```

Build:

```bash
./scripts/build.sh
```

## Najważniejsze endpointy

### Auth fotografa

```txt
GET  /api/auth/google/login
GET  /api/auth/google/callback
POST /api/auth/logout
GET  /api/me/profile
PUT  /api/me/profile
```

### Sesje fotografa

```txt
GET  /api/sessions
POST /api/sessions
GET  /api/sessions/{sessionId}
POST /api/sessions/{sessionId}/close
POST /api/sessions/{sessionId}/access/regenerate
```

### Upload zdjęć source

```txt
POST /api/sessions/{sessionId}/photos/presign
POST /api/sessions/{sessionId}/photos/{photoId}/complete
```

### Klient

```txt
GET  /api/client/access/by-token/{token}
POST /api/client/access/by-code
GET  /api/client/session/{sessionId}/photos
GET  /api/client/photos/{photoId}/proof-url
PUT  /api/client/session/{sessionId}/selections
POST /api/client/session/{sessionId}/submit
GET  /api/client/session/{sessionId}/download
```

### Płatność i finale

```txt
POST /api/sessions/{sessionId}/payment/mark-paid
POST /api/sessions/{sessionId}/finals/presign
POST /api/sessions/{sessionId}/finals/{finalId}/complete
POST /api/sessions/{sessionId}/deliveries/generate-zip
```

### Portfolio

```txt
GET  /api/public/galleries/featured?limit=4
GET  /api/public/photographers/{username}
GET  /api/public/photographers/{username}/galleries/{slug}
GET  /api/galleries
POST /api/galleries
GET  /api/galleries/{galleryId}
PUT  /api/galleries/{galleryId}
DELETE /api/galleries/{galleryId}
POST /api/galleries/{galleryId}/photos/presign
POST /api/galleries/{galleryId}/photos/{photoId}/complete
DELETE /api/galleries/{galleryId}/photos/{photoId}
```

## Statusy sesji

```txt
draft                 - sesja utworzona, można dodawać zdjęcia
processing            - trwa przygotowanie miniaturek/proofów
selecting             - klient wybiera zdjęcia
waiting_for_payment   - wybór zatwierdzony, oczekiwanie na płatność
editing               - fotograf przygotowuje finalne zdjęcia
delivered             - ZIP gotowy do pobrania
closed                - sesja zamknięta
archived              - sesja archiwalna
failed                - błąd przetwarzania
```

## Bezpieczeństwo

Najważniejsze zasady:

- Klient nie ma konta; dostęp jest przez kod/link.
- W bazie nie trzymamy plaintext kodu ani tokenu, tylko HMAC.
- Fotograf ma JWT w HttpOnly cookie.
- Klient po wejściu kodem/linkiem dostaje osobne HttpOnly cookie.
- Bucket storage jest prywatny.
- Dostęp do plików idzie przez signed/presigned URL-e.
- CORS musi mieć konkretny `FRONTEND_ORIGIN`, nie `*`, bo używane są cookies.
- Ownership guard dla fotografa ma zwracać `404` dla cudzych lub nieistniejących zasobów.

## follow-upy

Nieblokujące MVP:

- automatyczne e-maile do klienta,
- cleanup/retencja zamkniętych sesji,
- dodatkowe business eventy Mini-SIEM,
- lepsze recovery jobów `running`,
- FE/BE performance przy bardzo dużych sesjach,
- bramka płatności online,
- pełniejsze wsparcie RAW.
