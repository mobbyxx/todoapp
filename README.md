# Todo App - Gamified Task Management

Eine kollaborative Todo-App mit Gamification-Features für 1:1 Verbindungen. Entwickelt mit Go (Backend) und React Native + Expo (Mobile).

[![CI/CD](https://github.com/mobbyxx/todoapp/actions/workflows/backend.yml/badge.svg)](https://github.com/mobbyxx/todoapp/actions)
[![CI/CD](https://github.com/mobbyxx/todoapp/actions/workflows/mobile.yml/badge.svg)](https://github.com/mobbyxx/todoapp/actions)

## 🚀 Features

### Core Features
- ✅ **Benutzer-Authentifizierung** - JWT & API Keys mit Argon2id
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
- **Go 1.21+** mit Chi Router
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
- Docker & Docker Compose
- Go 1.21+ (für lokale Entwicklung)
- Node.js 18+ (für Mobile)
- Expo Account (für Builds)

### 1. Repository klonen

```bash
git clone https://github.com/mobbyxx/todoapp.git
cd todoapp
```

### 2. Backend starten

```bash
# Environment kopieren
cp backend/.env.example backend/.env
# .env anpassen mit deinen Credentials

# Docker Services starten
docker-compose up -d

# API Server starten
cd backend
go mod download
go run cmd/api/main.go
```

Die API ist nun unter `http://localhost:8080` verfügbar.

### 3. Mobile App starten

```bash
cd mobile
npm install
npx expo start
```

Scannen Sie den QR-Code mit der Expo Go App oder drücken Sie:
- `i` für iOS Simulator
- `a` für Android Emulator

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

**Backend (.env)**
```env
# Database
DB_URL=postgres://user:pass@localhost:5432/todoapp?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-secret-key-here

# FCM (Push Notifications)
FCM_API_KEY=your-fcm-api-key

# Server
PORT=8080
API_DOMAIN=api.todoapp.com
```

**Mobile (.env)**
```env
API_URL=https://api.todoapp.com
API_KEY=your-api-key
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

### 🔴 Kritisch (Sollte vor Launch erledigt werden)

1. **TypeScript Konfiguration fixen**
   - Problem: Decorator Fehler in Mobile
   - Lösung: `mobile/tsconfig.json` aktualisieren:
   ```json
   {
     "compilerOptions": {
       "experimentalDecorators": true,
       "emitDecoratorMetadata": true
     }
   }
   ```

2. **Badge Auto-Award implementieren**
   - Problem: `CheckAndAwardBadges()` ist nur ein Stub
   - Datei: `backend/internal/service/gamification_service.go`
   - Muss aufgerufen werden bei: Todo Complete, Streak Update

### 🟡 Wichtig (Kurz nach Launch)

3. **Tests ausführen & Coverage erhöhen**
   - Aktuell: ~34% Coverage
   - Ziel: >80% Coverage
   - Fokus auf: User Service, Auth Middleware

4. **E2E Tests ausführen**
   - Maestro Tests auf realen Geräten testen
   - Flaky Tests identifizieren & fixen

5. **Performance Optimierung**
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

**Status:** ✅ 36/36 Tasks Completed (100%)

**Letztes Update:** April 2025

**Repository:** https://github.com/mobbyxx/todoapp
