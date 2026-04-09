# Todo App - Gamified Task Management

Eine kollaborative Todo-App mit Gamification-Features für 1:1 Verbindungen. Entwickelt mit Go (Backend) und React Native + Expo (Mobile).

[![CI/CD](https://github.com/mobbyxx/todoapp/actions/workflows/backend.yml/badge.svg)](https://github.com/mobbyxx/todoapp/actions)
[![CI/CD](https://github.com/mobbyxx/todoapp/actions/workflows/mobile.yml/badge.svg)](https://github.com/mobbyxx/todoapp/actions)

## 🚀 Features

### Core Features
- ✅ **Benutzer-Authentifizierung** - JWT mit Argon2id (API Key Service existiert, aber nicht in Runtime verdrahtet — `noopAPIKeyService` aktiv, kein Repository implementiert)
- ✅ **1:1 Verbindungen** - Einladungslinks & QR-Code Verbindungen
- ✅ **Geteilte Todos** - CRUD mit optimistischem Locking
- ✅ **Offline-First** - WatermelonDB mit Delta-Sync
- ✅ **Konfliktlösung** - Per-Feld Strategien

### Gamification
- 🏆 **XP & Level System** - 8 Level (0-12,000 XP)
- 🎯 **Badges** - 5 Badges (First Todo, Week Warrior, etc.)
- 🎁 **Custom Rewards** - Erstellen & Einlösen
- 🎯 **Shared Goals** - Gemeinsame Ziele mit Partner
- 📊 **Streaks** - Login-Streaks mit Freeze-Tokens

### Technische Highlights
- 🔒 **Security** - SSL A+, Rate Limiting, Security Headers
- 📱 **Push Notifications** - FCM/APNS mit Worker
- 🔄 **Sync** - Hybrid Logical Clocks (HLC)
- 🛡️ **Anti-Cheat** - Serverseitige Validierung
- 📊 **Observability** - Prometheus Metrics & Health Checks

## 📁 Projektstruktur

```
todoapp/
├── backend/                 # Go REST API
│   ├── cmd/
│   │   ├── api/            # API Server
│   │   └── worker/         # Notification Worker
│   ├── internal/
│   │   ├── domain/         # Entities & Interfaces
│   │   ├── repository/     # Data Access (PostgreSQL)
│   │   ├── service/        # Business Logic
│   │   ├── handler/        # HTTP Handlers
│   │   └── middleware/     # Auth, Logging, Security
│   ├── migrations/         # 23 SQL Migrations
│   └── tests/              # Integration & Benchmarks
├── mobile/                  # React Native + Expo
│   ├── app/                # Expo Router Screens
│   ├── components/         # UI Components
│   ├── services/           # API & Database
│   ├── models/             # WatermelonDB Models
│   ├── hooks/              # React Hooks
│   └── tests/              # Maestro E2E Tests
├── nginx/                   # Reverse Proxy & SSL
├── .github/workflows/       # CI/CD Pipelines
└── docs/                    # Documentation
```

## 🛠️ Technologie-Stack

### Backend
- **Go 1.23+** mit Chi Router
- **PostgreSQL 15** - Primäre Datenbank
- **Redis 7** - Caching & Sessions
- **JWT** - Authentication
- **Testcontainers** - Integration Tests
- **Prometheus** - Metrics

### Mobile
- **React Native** mit Expo SDK 52
- **TypeScript** - Type Safety
- **Expo Router** - File-based Navigation
- **WatermelonDB** - Offline-First Database
- **TanStack Query** - Server State Management
- **Zustand** - UI State Management
- **Reanimated** - Animationen

### Infrastructure
- **Docker & Docker Compose**
- **Nginx** - Reverse Proxy mit SSL
- **GitHub Actions** - CI/CD
- **Let's Encrypt** - SSL Zertifikate

## 🚀 Quick Start

### Voraussetzungen

| Komponente | Minimum | Empfohlen |
|---|---|---|
| Docker & Docker Compose | v20+ | v24+ |
| Go | 1.23+ | 1.23+ |
| Node.js | 18+ | 20 LTS |
| yarn | 4.x (via corepack) | 4.x |
| Expo Go App | — | Auf dem Handy installieren |

### 1. Repository klonen

```bash
git clone https://github.com/mobbyxx/todoapp.git
cd todoapp
```

### 2. Backend starten

```bash
# Docker Services (PostgreSQL + Redis)
docker-compose up -d postgres redis

# Warten bis postgres bereit ist
sleep 3

# Backend starten
cd backend
go mod download

# Pflicht-Umgebungsvariablen:
export DB_URL="postgres://todoapp:todoapp@localhost:5432/todoapp?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
export JWT_SECRET="change-me-in-production"
export PORT=8090

go run ./cmd/api
```

Die API ist nun unter `http://localhost:8090` verfügbar.  
Health Check: `curl http://localhost:8090/health/live`

### 3. Mobile App starten

> **⚠️ Bekanntes Problem:** `package.json` hat React Native 0.81 + React 19, aber Expo SDK 52 erwartet RN 0.76 + React 18. Vor dem Start muss auf **Expo SDK 53** aktualisiert werden (unterstützt RN 0.81 + React 19).

```bash
cd mobile

# Abhängigkeiten installieren
corepack enable
yarn install

# TypeScript prüfen (muss fehlerfrei sein)
npx tsc --noEmit

# Expo SDK 53 Upgrade (einmalig nötig!)
npx expo install expo@^53 --fix

# Starten
npx expo start
```

**Auf dem iPhone/Android:**
1. [Expo Go](https://expo.dev/go) aus dem App Store installieren
2. QR-Code scannen (iOS: Kamera-App, Android: Expo Go direkt)
3. Handy und Computer müssen im selben WLAN sein

**Im Simulator:**
- `i` → iOS Simulator (macOS + Xcode nötig)
- `a` → Android Emulator (Android Studio nötig)

**API-Verbindung konfigurieren:**
Die Mobile-App verbindet sich standardmäßig mit der API-URL aus `mobile/services/api.ts`. Für lokale Entwicklung die eigene IP einsetzen (nicht `localhost` — das ist auf dem Handy das Handy selbst):
```bash
# Eigene IP finden (macOS)
ifconfig | grep "inet " | grep -v 127.0.0.1

# Dann in mobile/services/api.ts:
# API_URL = "http://192.168.x.x:8090"
```

## 🧪 Testing

### Backend Tests

```bash
cd backend

# Unit Tests
go test ./... -v

# Mit Coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Integration Tests (erfordert Docker)
make test-integration

# Benchmarks
make benchmark

# Security Scan
make test-security
```

### Mobile Tests

```bash
cd mobile

# TypeScript Check
npx tsc --noEmit

# ESLint
npm run lint

# E2E Tests mit Maestro
maestro test tests/onboarding.yaml
maestro test tests/todo-flow.yaml
maestro test tests/auth-flow.yaml
```

### Load Testing

```bash
# k6 Load Tests (erfordert laufendes Backend)
cd tests/load
k6 run load_test.js
```

## 📱 App Stores

### iOS App Store

```bash
cd mobile

# Production Build
eas build --platform ios --profile production

# Oder mit automatischem Submit
eas build --platform ios --profile production --auto-submit
```

### Google Play Store

```bash
cd mobile

# Production Build
eas build --platform android --profile production

# Mit automatischem Submit
eas build --platform android --profile production --auto-submit
```

## 🚀 Deployment

### Option 1: Self-Hosted (Empfohlen für MVP)

1. **Server vorbereiten** (z.B. Hetzner, DigitalOcean)
```bash
# Auf dem Server
git clone https://github.com/mobbyxx/todoapp.git
cd todoapp
```

2. **SSL Zertifikate einrichten**
```bash
./scripts/init-letsencrypt.sh
```

3. **Production starten**
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Option 2: Cloud Deployment

**Backend:**
- Railway, Render, oder Fly.io
- PostgreSQL: Supabase oder AWS RDS
- Redis: Upstash oder AWS ElastiCache

**Mobile:**
- Expo Application Services (EAS)
- OTA Updates via Expo

## 📊 Monitoring

### Health Checks
- Liveness: `GET /health/live`
- Readiness: `GET /health/ready`
- Metrics: `GET /metrics` (Prometheus)

### Logs
```bash
# Backend Logs
docker-compose logs -f api

# Worker Logs
docker-compose logs -f worker

# Nginx Logs
docker-compose logs -f nginx
```

## 🔧 Konfiguration

### Wichtige Umgebungsvariablen

**Backend**
```env
# Database (Pflicht)
DB_URL=postgres://todoapp:todoapp@localhost:5432/todoapp?sslmode=disable

# Redis (Pflicht)
REDIS_URL=redis://localhost:6379

# JWT (Pflicht)
JWT_SECRET=change-me-in-production

# Server
PORT=8090
API_DOMAIN=api.todoapp.com

# FCM (Optional — Push Notifications)
FCM_API_KEY=your-fcm-api-key
```

**Mobile**  
API-URL wird in `mobile/services/api.ts` gesetzt. Für lokale Entwicklung die eigene LAN-IP verwenden:
```
http://192.168.x.x:8090
```

## 📝 API Dokumentation

Die vollständige API Dokumentation ist verfügbar unter:
- **OpenAPI Spec:** `backend/docs/openapi.yaml`
- **Swagger UI:** `http://localhost:8080/swagger` (nach Start)

### Wichtige Endpoints

```
POST   /api/v1/auth/register          # Registrierung
POST   /api/v1/auth/login             # Login
POST   /api/v1/connections/invite     # Einladung erstellen
GET    /api/v1/todos                  # Todos listen
POST   /api/v1/todos                  # Todo erstellen
POST   /api/v1/todos/:id/complete     # Todo abschließen
GET    /api/v1/users/me/stats         # Gamification Stats
POST   /api/v1/rewards                # Reward erstellen
POST   /api/v1/sync                   # Delta Sync
```

## 🎯 Nächste Schritte

### ✅ Erledigt

1. ~~**TypeScript Konfiguration fixen**~~ — `experimentalDecorators` + `emitDecoratorMetadata` hinzugefügt
2. ~~**Badge Auto-Award implementieren**~~ — Vollständige Implementierung inkl. Wiring in Todo-Completion-Flow
3. ~~**Alle Backend-Tests fixen**~~ — 31 Testfehler behoben, `go test ./...` grün

### 🟡 Wichtig (Vor Launch)

4. **Test Coverage erhöhen**
   - Aktuell: ~30% Coverage
   - Ziel: >80% Coverage
   - Fehlend: Repository-Layer, Main-Entrypoints, Domain-Utilities

5. **Mobile: Expo SDK 53 Upgrade + Unit Tests schreiben**
   - `yarn install` + `tsc --noEmit` bestanden ✅
   - Expo SDK 52 → 53 nötig (RN 0.81 + React 19 Support)
   - Aktuell: 0 Jest/Testing-Library Tests
   - Vorhanden: 5 Maestro E2E Specs (YAML)
   - Metro Bundler startet noch nicht (SDK-Version-Mismatch)

6. **Badge-Vergabe bei Streak-Update wiren**
   - `CheckAndAwardBadges` wird bei Todo-Completion aufgerufen
   - Fehlt noch: Aufruf bei Streak-Updates (für streak-basierte Badges)

7. **CI/CD Secrets konfigurieren**
   - 21 GitHub Secrets benötigt (DB, Redis, JWT, Expo, Store Credentials)
   - SSL-Zertifikate + DNS (api.todo.com, staging-api.todo.com)

8. **Performance Optimierung**
   - API Response Time < 200ms (p95)
   - Mobile App Start < 3s
   - Bundle Size analysieren

### 🟢 Erweiterungen (Zukünftige Features)

6. **Dark Mode**
   - Theme Context implementieren
   - Farbschema erstellen

7. **Statistiken & Insights**
   - Todo Completion Rate
   - Weekly/Monthly Reports
   - Productivity Trends

8. **Erinnerungen**
   - Lokale Notifications
   - Push für Deadline-Erinnerungen

9. **Export/Backup**
   - JSON Export
   - Cloud Backup

10. **Social Features**
    - Activity Feed (nur für Partner)
    - Nudges (freundliche Erinnerungen)

## 🤝 Mitwirken

1. Fork das Repository
2. Erstelle einen Feature Branch (`git checkout -b feature/amazing-feature`)
3. Committe deine Änderungen (`git commit -m 'feat: add amazing feature'`)
4. Pushe zum Branch (`git push origin feature/amazing-feature`)
5. Öffne einen Pull Request

### Commit Convention

Wir folgen [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - Neue Features
- `fix:` - Bug Fixes
- `docs:` - Dokumentation
- `test:` - Tests
- `refactor:` - Code Refactoring
- `security:` - Security Fixes
- `perf:` - Performance Improvements

## 📄 Lizenz

Dieses Projekt ist unter der MIT Lizenz lizenziert. Siehe [LICENSE](LICENSE) für Details.

## 🙏 Danksagung

- [Go](https://golang.org/) - Backend Language
- [Expo](https://expo.dev/) - Mobile Framework
- [WatermelonDB](https://github.com/Nozbe/WatermelonDB) - Offline Database
- [Chi](https://github.com/go-chi/chi) - HTTP Router

---

**Status:** Backend stabil — Tests grün (31 Fixes), 25/25 API-Endpoints getestet, Badge-System implementiert + in Todo-Completion gewired, Routing + Schema-Bugs behoben. Mobile: TypeScript kompiliert fehlerfrei, Expo SDK 52→53 Upgrade ausstehend für Metro-Start. Verbleibend: Test Coverage (~30% → 80%), Expo SDK Upgrade, Mobile Unit Tests, CI/CD Secrets.

**Letztes Update:** April 2026

**Repository:** https://github.com/mobbyxx/todoapp
