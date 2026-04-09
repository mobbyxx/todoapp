# Work Plan: Kollaborative Todo-App mit Gamification

## TL;DR

> **Quick Summary**: Eine mobile Todo-App für 1:1 Zusammenarbeit mit Gamification-Elementen. React Native (Expo) Frontend, Go Backend mit PostgreSQL, skalierbare Clean Architecture.
>
> **Deliverables**:
> - React Native Mobile App (iOS + Android)
> - Go REST API mit PostgreSQL
> - Gamification-System (Punkte, Level, Badges)
> - 1:1 User-Verbindungen via Einladungslinks & QR-Code
> - Push Notifications
> - API Keys für Automation
>
> **Estimated Effort**: Large (5+ Wochen)
> **Parallel Execution**: YES - 4 Wellen mit bis zu 8 parallelen Tasks
> **Critical Path**: DB Schema → Auth → API Core → Sync Engine → Gamification → Mobile UI → Integration

---

## Context

### Original Request
Eine kollaborative Todo-App für mehrere Personen, bei der man sich gegenseitig Todos schreiben kann, mit Gamification und individuell festlegbaren Belohnungen. Mit echter Mobile App, Account-Verbindung und zentraler API für Automatisierung.

### Interview Summary
**Key Discussions**:
- Mobile: React Native mit Expo (besser für REST APIs, einfacheres Hiring)
- Backend: Go + PostgreSQL + Redis
- Features: 1:1 Verbindungen, geteilte Todos, Gamification (Punkte/Level/Badges), Custom Rewards
- Verbindung: Einladungslinks UND QR-Code
- Sicherheit: Domain vorhanden, SSL möglich, DB-Verschlüsselung empfohlen
- Skalierung: <100 User initial, aber Architektur auf Wachstum ausgelegt

### Metis Review Findings

**Critical Gaps Identified** (addressed in this plan):
1. **Konfliktlösung**: WatermelonDB's Default-Verhalten führt zu Datenverlust - custom Resolver pro Feld-Typ implementiert
2. **Gamification-Formeln**: XP-Kurve, Level-Berechnung, Badge-Kriterien spezifiziert
3. **Anti-Cheat**: Server-side Validation für alle gamifizierten Aktionen
4. **Sync-Protokoll**: Delta-Sync mit Timestamps und Konflikt-Tracking
5. **Notification Queue**: PostgreSQL-basierte Queue für zuverlässige Zustellung

**Security Directives**:
- JWT mit 15-min Ablauf + Refresh Tokens
- API Keys mit bcrypt Hashing
- Rate Limiting (100 req/min default, 5 req/min auth)
- Security Headers auf allen Responses
- SQL Injection Prevention (parameterized queries)

**Gamification Guardrails**:
- KEINE Leaderboards (sinnlos bei 2 Usern)
- Collaborative Goals statt Competition
- Streak Freeze Tokens (1/Monat, mehr er verdienen)
- Lineare Level-Kurve (keine exponentiellen Anforderungen)

---

## Work Objectives

### Core Objective
Eine produktionsreife mobile Todo-App für 1:1 Zusammenarbeit mit Gamification-Elementen, die offline funktioniert und zuverlässig synchronisiert.

### Concrete Deliverables

**Backend (Go)**:
- `cmd/api/main.go` - API Server
- `cmd/worker/main.go` - Background Notification Worker
- `internal/` - Clean Architecture (domain, repository, service, handler, middleware)
- REST API Endpoints für: Auth, Users, Connections, Todos, Gamification, Sync
- PostgreSQL Schema mit 17 Tabellen
- JWT + API Key Authentication
- Push Notification Queue
- Anti-Cheat Validation

**Mobile (React Native + Expo)**:
- Expo SDK 52 Projekt
- TanStack Query für Server-State
- Zustand für UI-State
- WatermelonDB für Offline-Storage
- Expo Notifications für Push
- Deep Linking für Einladungen
- QR-Code Scanner
- Reanimated für Gamification-Animationen

**Infrastructure**:
- Docker Compose Setup
- PostgreSQL + Redis
- Nginx Reverse Proxy
- SSL/TLS Konfiguration
- Monitoring & Logging

### Definition of Done
- [ ] Alle API Endpoints dokumentiert (OpenAPI/Swagger)
- [ ] Mobile App auf iOS und Android getestet
- [ ] Sync funktioniert offline → online
- [ ] Push Notifications zuverlässig
- [ ] Gamification-System ohne Exploits
- [ ] Security Audit bestanden
- [ ] Load Test: 100+ gleichzeitige User
- [ ] CI/CD Pipeline läuft

### Must Have
- [ ] User Registration/Login mit JWT
- [ ] 1:1 Verbindungen via Einladungslinks & QR-Code
- [ ] Geteilte Todos (CRUD, Status-Updates)
- [ ] Offline-First mit Sync
- [ ] Gamification: Punkte, Level (1-8), Badges
- [ ] Custom Rewards (erstellen, einlösen)
- [ ] Push Notifications
- [ ] API Keys für externe Automation
- [ ] Konfliktlösung bei Sync-Konflikten

### Must NOT Have (Guardrails)
- [ ] Keine Leaderboards (sinnlos bei 2 Usern)
- [ ] Keine Gruppen/Teams (nur 1:1)
- [ ] Keine Datei-Uploads
- [ ] Keine Echtzeit-Chat-Funktion
- [ ] Keine Web-Version (nur Mobile)
- [ ] Keine Social Media Integration
- [ ] Keine KI-Features im MVP

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: NO (wird aufgesetzt)
- **Automated tests**: TDD für Business Logic, Integration Tests für API
- **Framework**: 
  - Backend: `go test` mit Testcontainers
  - Mobile: Jest + React Native Testing Library + Maestro E2E
- **If TDD**: Jede Task mit "Business Logic" hat Unit Tests als Teil der Acceptance Criteria

### QA Policy
Jede Task hat Agent-Executable QA Scenarios:

**Backend Tasks**:
- `curl` Tests für API Endpoints
- Datenbank-Assertions für Schema
- Integration Tests mit Testcontainers

**Mobile Tasks**:
- Jest Tests für Components/Logic
- Maestro E2E für kritische Flows
- Build-Validation für iOS/Android

**Infrastructure Tasks**:
- Docker Compose up → health checks
- SSL-Zertifikat Validierung
- Monitoring Dashboard erreichbar

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation - Woche 1):
├── Task 1.1: Projekt-Setup & Docker Infrastructure
├── Task 1.2: PostgreSQL Schema (17 Tabellen)
├── Task 1.3: Go Projekt-Struktur & Dependencies
├── Task 1.4: JWT Authentication System
├── Task 1.5: API Key Management
├── Task 1.6: Expo Mobile Projekt Setup
├── Task 1.7: Security Middleware Stack
└── Task 1.8: Basis API Client (Mobile)

Wave 2 (Core Features - Woche 2-3):
├── Task 2.1: User Management API
├── Task 2.2: 1:1 Connections (Einladungen)
├── Task 2.3: QR-Code Verbindung
├── Task 2.4: Todo CRUD API
├── Task 2.5: Sync Engine (Delta Sync)
├── Task 2.6: Konfliktlösung
├── Task 2.7: Mobile Auth Flow
├── Task 2.8: Mobile Todo UI

Wave 3 (Gamification - Woche 4):
├── Task 3.1: Gamification Domain Model
├── Task 3.2: XP & Level System
├── Task 3.3: Badge System
├── Task 3.4: Anti-Cheat Validation
├── Task 3.5: Custom Rewards
├── Task 3.6: Gamification UI
├── Task 3.7: Collaborative Goals
└── Task 3.8: Animationen & Celebration

Wave 4 (Integration & Production - Woche 5):
├── Task 4.1: Push Notification Queue
├── Task 4.2: Notification Worker
├── Task 4.3: Deep Linking (Einladungen)
├── Task 4.4: Offline-First Sync
├── Task 4.5: Observability & Monitoring
├── Task 4.6: CI/CD Pipeline
├── Task 4.7: E2E Tests & Performance
└── Task 4.8: Security Hardening

Wave FINAL (Review & Launch):
├── Task F1: Plan Compliance Audit (oracle)
├── Task F2: Code Quality Review (unspecified-high)
├── Task F3: Real Manual QA (unspecified-high)
└── Task F4: Security Audit (deep)
```

### Dependency Matrix

- **Wave 1 Tasks**: keine Dependencies, können alle parallel starten
- **Wave 2 Tasks**: brauchen Wave 1 (Auth, Schema, Setup)
- **Wave 3 Tasks**: brauchen Wave 2 (Todos, Connections)
- **Wave 4 Tasks**: brauchen Wave 3 (Gamification) + Wave 2 (Sync)
- **FINAL Tasks**: brauchen alle Implementation Tasks

### Agent Dispatch Summary

- **Wave 1**: 
  - T1.1, T1.3, T1.7 → `quick` (Setup, Konfiguration)
  - T1.2, T1.4, T1.5 → `deep` (Schema, Auth, Security)
  - T1.6, T1.8 → `unspecified-high` (Mobile Setup)

- **Wave 2**:
  - T2.1, T2.2, T2.3 → `unspecified-high` (API Features)
  - T2.4, T2.5, T2.6 → `deep` (Sync Engine, Konflikte)
  - T2.7, T2.8 → `visual-engineering` (Mobile UI)

- **Wave 3**:
  - T3.1, T3.2, T3.3, T3.4 → `deep` (Gamification Logic)
  - T3.5, T3.6, T3.7, T3.8 → `unspecified-high` (Features + UI)

- **Wave 4**:
  - T4.1, T4.2, T4.3 → `unspecified-high` (Integration)
  - T4.4 → `deep` (Offline-Sync)
  - T4.5, T4.6 → `quick` (DevOps)
  - T4.7, T4.8 → `unspecified-high` (Testing + Security)

---

## TODOs

- [x] 1.1. Projekt-Setup & Docker Infrastructure

  **What to do**:
  - Erstelle Projekt-Root mit `backend/` und `mobile/` Ordnern
  - Erstelle `docker-compose.yml` mit PostgreSQL 15, Redis 7, Nginx
  - Konfiguriere `Dockerfile` für Go Backend (multi-stage build)
  - Erstelle `.env.example` mit allen benötigten Umgebungsvariablen
  - Setup Ordner-Struktur für Migrations und Configs
  - Erstelle `Makefile` mit nützlichen Commands (build, test, migrate, etc.)

  **Must NOT do**:
  - Keine Go-Code schreiben (nur Setup)
  - Keine Mobile App Code
  - Keine SSL-Konfiguration in diesem Task

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Setup und Konfiguration, keine komplexe Logik
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (mit T1.2, T1.3, T1.4, T1.5, T1.6, T1.7, T1.8)
  - **Blocks**: T1.2, T1.3, T2.x
  - **Blocked By**: None

  **References**:
  - `docker-compose.yml` Beispiel: Metis Research (PostgreSQL + Redis + Nginx)
  - Go Multi-Stage Dockerfile: `golang:1.21-alpine` → `alpine:latest`
  - Ordnerstruktur: Clean Architecture Pattern (cmd/, internal/)

  **Acceptance Criteria**:
  - [ ] `docker-compose up` startet PostgreSQL, Redis, Nginx ohne Fehler
  - [ ] PostgreSQL ist erreichbar auf Port 5432
  - [ ] Redis ist erreichbar auf Port 6379
  - [ ] `make build` erstellt Go Binary
  - [ ] `.env.example` enthält alle notwendigen Variablen
  - [ ] Ordner-Struktur existiert: `backend/cmd/api`, `backend/internal/`, `backend/migrations/`

  **QA Scenarios**:
  ```
  Scenario: Docker Compose startet alle Services
    Tool: Bash
    Preconditions: Docker Desktop läuft
    Steps:
      1. cd backend && docker-compose up -d
      2. sleep 10
      3. docker-compose ps
    Expected Result: Alle 3 Container zeigen Status "Up"
    Evidence: Screenshot von docker-compose ps
  ```

  **Commit**: YES
  - Message: `chore(setup): initial docker infrastructure`
  - Files: `docker-compose.yml`, `Dockerfile`, `.env.example`, `Makefile`

- [x] 1.2. PostgreSQL Schema (17 Tabellen)

  **What to do**:
  - Erstelle Migrations für alle 17 Tabellen:
    1. `users` - mit Auth, Profil, Gamification-Daten
    2. `connections` - 1:1 Beziehungen
    3. `todos` - mit Version für Optimistic Locking
    4. `points_history` - Gamification Ledger
    5. `badges` - Badge Definitionen
    6. `user_badges` - Verknüpfung
    7. `levels` - Level Definitionen
    8. `rewards` - Custom Rewards
    9. `reward_redemptions` - Einlösungen
    10. `push_notification_tokens` - FCM/APNS Tokens
    11. `api_keys` - API Key Management
    12. `notifications` - In-App Notifications
    13. `sync_conflicts` - Konflikt-Tracking
    14. `notification_queue` - Push Queue
    15. `audit_log` - Audit Trail
    16. `rate_limits` - Rate Limiting
    17. `shared_goals` - Collaborative Goals
  - Erstelle alle Indizes (B-tree, Partial, GIN für JSONB)
  - Erstelle Trigger für `updated_at`
  - Erstelle Seed-Daten für Levels und Default Badges
  - Verwende `pressly/goose/v3` für Migrations

  **Must NOT do**:
  - Keine Business Logic in Migrations
  - Keine Datenbank-Code außerhalb von `migrations/`

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Komplexes Schema mit vielen Beziehungen und Constraints
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T1.3, T2.x
  - **Blocked By**: T1.1 (Docker muss laufen für Tests)

  **References**:
  - PostgreSQL Schema Design: Librarian Research (12 Tabellen + 5 neue)
  - Goose Migrations: `pressly/goose/v3`
  - Indexing Strategy: B-tree, Partial, GIN für JSONB

  **Acceptance Criteria**:
  - [ ] Alle 17 Migration-Files in `migrations/` Ordner
  - [ ] `goose up` führt alle Migrations erfolgreich aus
  - [ ] `goose status` zeigt alle als "Applied"
  - [ ] Alle Foreign Keys und Constraints definiert
  - [ ] Seed-Daten für Levels 1-8 und Default Badges
  - [ ] Test: Datenbank-Schema lässt sich auf leerer DB aufbauen

  **QA Scenarios**:
  ```
  Scenario: Migration läuft erfolgreich
    Tool: Bash
    Preconditions: PostgreSQL Container läuft
    Steps:
      1. cd backend && goose postgres "postgresql://user:pass@localhost:5432/todoapp?sslmode=disable" up
      2. goose status
      3. psql -c "\dt" todoapp
    Expected Result: Alle 17 Tabellen werden aufgelistet
    Evidence: Terminal Output mit Tabellen-Liste
  ```

  **Commit**: YES
  - Message: `feat(db): add complete schema with 17 tables`
  - Files: `migrations/*.sql`

- [x] 1.3. Go Projekt-Struktur & Dependencies

  **What to do**:
  - Initialisiere Go Module: `go mod init github.com/user/todo-api`
  - Erstelle Clean Architecture Ordnerstruktur:
    - `internal/domain/` - Entities, Interfaces
    - `internal/repository/` - Data Access
    - `internal/service/` - Business Logic
    - `internal/handler/` - HTTP Handlers
    - `internal/middleware/` - Auth, Logging, etc.
    - `internal/infrastructure/` - Push, Cache
  - Installiere Dependencies:
    - `go-chi/chi/v5` - Router
    - `jackc/pgx/v5` - PostgreSQL Driver
    - `redis/go-redis/v9` - Redis Client
    - `golang-jwt/jwt/v5` - JWT
    - `go-playground/validator/v10` - Validation
    - `google/uuid` - UUIDs
    - `rs/zerolog` - Logging
    - `github.com/prometheus/client_golang/prometheus` - Metrics
  - Erstelle `config/config.go` für Umgebungsvariablen
  - Erstelle Basis Error Types in `internal/domain/errors.go`

  **Must NOT do**:
  - Keine Handler implementieren (nur Interfaces)
  - Keine Business Logic

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Setup und Boilerplate
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T1.4, T1.5, T1.7, T2.x
  - **Blocked By**: T1.1

  **References**:
  - Clean Architecture: Standard Go Projekt-Layout
  - Dependencies: Librarian Research (Go Backend Patterns)

  **Acceptance Criteria**:
  - [ ] `go.mod` und `go.sum` existieren
  - [ ] `go build ./...` kompiliert ohne Fehler
  - [ ] Alle Dependencies in go.mod aufgelistet
  - [ ] Ordnerstruktur: `internal/domain/`, `repository/`, `service/`, `handler/`, `middleware/`, `infrastructure/`
  - [ ] `config.go` lädt alle Umgebungsvariablen
  - [ ] Basis Error Types definiert

  **QA Scenarios**:
  ```
  Scenario: Go Projekt baut erfolgreich
    Tool: Bash
    Steps:
      1. cd backend
      2. go mod tidy
      3. go build ./...
    Expected Result: Build erfolgreich, keine Fehler
    Evidence: Terminal Output "Build successful"
  ```

  **Commit**: YES
  - Message: `chore(backend): setup go project structure and dependencies`
  - Files: `go.mod`, `go.sum`, `internal/**/`, `config/`

- [x] 1.4. JWT Authentication System

  **What to do**:
  - Erstelle JWT Service in `internal/service/jwt_service.go`:
    - `GenerateTokenPair(userID string)` → Access Token (15min) + Refresh Token (7 Tage)
    - `ValidateAccessToken(token string)` → CustomClaims
    - `ValidateRefreshToken(token string)` → New Token Pair
    - `RevokeToken(token string)` → Blacklist in Redis
  - Implementiere Token Versioning für Logout-all
  - Erstelle Custom Claims Struktur mit UserID, Roles, TokenVersion
  - Verwende `golang-jwt/jwt/v5` mit HS256 (symmetric)
  - Speichere JWT Secret in Umgebungsvariable
  - Unit Tests für alle Methoden

  **Must NOT do**:
  - Keine HTTP Handler (nur Service Layer)
  - Keine User Registration (nur Token-Management)

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Security-kritisch, muss korrekt implementiert werden
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T1.7, T2.1, T2.7
  - **Blocked By**: T1.3

  **References**:
  - JWT Patterns: Librarian Research (Access/Refresh Tokens, Token Versioning)
  - Library: `golang-jwt/jwt/v5`
  - Security: Short-lived access tokens (15min), refresh token rotation

  **Acceptance Criteria**:
  - [ ] JWT Service mit allen Methoden implementiert
  - [ ] Access Token: 15 Minuten TTL
  - [ ] Refresh Token: 7 Tage TTL
  - [ ] Token Versioning für Logout-all
  - [ ] Redis Blacklist für revoked tokens
  - [ ] Unit Tests: 100% Coverage für JWT Service
  - [ ] Test: Token Generation → Validation → Refresh → Revocation

  **QA Scenarios**:
  ```
  Scenario: JWT Token Lifecycle
    Tool: Bash (Go Test)
    Steps:
      1. go test -v ./internal/service/... -run TestJWT
    Expected Result: Alle Tests PASS, 100% Coverage
    Evidence: Test Output mit Coverage Report
  ```

  **Commit**: YES
  - Message: `feat(auth): implement JWT service with access/refresh tokens`
  - Files: `internal/service/jwt_service.go`, `internal/service/jwt_service_test.go`

- [x] 1.5. API Key Management

  **What to do**:
  - Erstelle API Key Service in `internal/service/apikey_service.go`:
    - `GenerateAPIKey(userID, name, scopes)` → Returns `fullKey` (prefix+random) + speichert Hash
    - `ValidateAPIKey(key)` → Prüft Hash, prüft Ablauf, aktualisiert last_used_at
    - `RevokeAPIKey(keyID)` → Deaktiviert Key
    - `ListAPIKeys(userID)` → Alle Keys des Users
  - Format: `{prefix}_{version}_{base64url(random32bytes)}`
  - Hashing mit Argon2id (`golang.org/x/crypto/argon2`)
  - Speichere: key_hash, key_prefix (erste 8 Zeichen), scopes, expires_at
  - Scopes als Array: `["read:todos", "write:todos"]`
  - Unit Tests für alle Methoden

  **Must NOT do**:
  - Keine Rate Limiting in diesem Task
  - Keine HTTP Handler

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Security-kritisch (API Keys = privilegierter Zugang)
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T2.x (API Endpoints)
  - **Blocked By**: T1.3

  **References**:
  - API Key Pattern: Librarian Research (Argon2id Hashing, Prefix+Hash)
  - Library: `golang.org/x/crypto/argon2`
  - Security: Keys niemals plain speichern, nur Hash

  **Acceptance Criteria**:
  - [ ] API Key Service implementiert
  - [ ] Key Format: `ouk_v1_ABC123xyz...`
  - [ ] Argon2id Hashing für Storage
  - [ ] Scope-basierte Validierung
  - [ ] Ablauf-Prüfung (expires_at)
  - [ ] Unit Tests: 100% Coverage
  - [ ] Test: Generate → Validate → Revoke Flow

  **QA Scenarios**:
  ```
  Scenario: API Key Lifecycle
    Tool: Bash (Go Test)
    Steps:
      1. go test -v ./internal/service/... -run TestAPIKey
    Expected Result: Alle Tests PASS, Key Hash korrekt, Scopes validiert
    Evidence: Test Output
  ```

  **Commit**: YES
  - Message: `feat(auth): implement API key management with argon2 hashing`
  - Files: `internal/service/apikey_service.go`, `internal/service/apikey_service_test.go`

- [x] 1.6. Expo Mobile Projekt Setup

  **What to do**:
  - Initialisiere Expo Projekt: `npx create-expo-app mobile --template blank-typescript`
  - Installiere Dependencies:
    - `expo` ~52
    - `expo-router` ~4 (File-based Navigation)
    - `@tanstack/react-query` ^5
    - `zustand` ^5
    - `@nozbe/watermelondb` ^0.27
    - `expo-notifications` ~0.29
    - `expo-camera` ~16 (für QR-Code)
    - `expo-linking` ~7 (Deep Linking)
    - `react-native-reanimated` ^4
    - `lottie-react-native` ^7
    - `@shopify/flash-list` ^1
  - Erstelle Ordnerstruktur:
    - `app/` - Expo Router Screens
    - `components/` - React Components
    - `hooks/` - Custom Hooks
    - `services/` - API Client, Database
    - `stores/` - Zustand Stores
    - `types/` - TypeScript Types
  - Konfiguriere `app.json` mit Deep Linking Scheme
  - Erstelle Basis TypeScript Types für API

  **Must NOT do**:
  - Keine UI Components implementieren
  - Keine API Calls
  - Keine Navigation Screens (nur Setup)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Mobile Setup, viele Dependencies zu koordinieren
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T1.8, T2.7, T2.8
  - **Blocked By**: None

  **References**:
  - Mobile Stack: Librarian Research (Expo SDK 52, TanStack Query, WatermelonDB)
  - Expo Router: File-based Navigation
  - Deep Linking: `expo-linking`

  **Acceptance Criteria**:
  - [ ] Expo Projekt läuft: `npx expo start` startet Metro Bundler
  - [ ] Alle Dependencies in `package.json`
  - [ ] Ordnerstruktur: `app/`, `components/`, `hooks/`, `services/`, `stores/`, `types/`
  - [ ] `app.json` konfiguriert mit Deep Linking Scheme
  - [ ] TypeScript Types für: User, Todo, Connection, Badge, etc.
  - [ ] `tsconfig.json` mit strict mode

  **QA Scenarios**:
  ```
  Scenario: Expo Projekt startet
    Tool: Bash
    Steps:
      1. cd mobile && npm install
      2. npx expo start --web
      3. curl http://localhost:8081
    Expected Result: Metro Bundler läuft, Web-Version erreichbar
    Evidence: Screenshot vom laufenden Expo Server
  ```

  **Commit**: YES
  - Message: `chore(mobile): setup expo project with dependencies`
  - Files: `mobile/package.json`, `mobile/app.json`, `mobile/tsconfig.json`, `mobile/app/`, `mobile/components/`, etc.

- [x] 1.7. Security Middleware Stack

  **What to do**:
  - Erstelle Middleware in `internal/middleware/`:
    - `auth.go` - JWT Validierung, setzt Claims in Context
    - `apikey.go` - API Key Validierung
    - `logging.go` - Structured Logging (zerolog)
    - `recovery.go` - Panic Recovery
    - `security.go` - Security Headers (HSTS, CSP, etc.)
    - `cors.go` - CORS Konfiguration
  - Verkette Middleware in korrekter Reihenfolge:
    1. Recovery (fängt Panics)
    2. RequestID (für Tracing)
    3. Logging (zeitliche Messung)
    4. Security Headers
    5. CORS
    6. Auth (JWT oder API Key)
  - Implementiere Auth Middleware die JWT ODER API Key akzeptiert
  - Security Headers: X-Content-Type-Options, X-Frame-Options, CSP, HSTS
  - CORS: Explicit origins, nicht `*` in Production

  **Must NOT do**:
  - Keine Rate Limiting (kommt später)
  - Keine HTTP Handler

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Security-kritisch, OWASP Compliance
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T2.x (alle API Endpoints)
  - **Blocked By**: T1.3, T1.4, T1.5

  **References**:
  - Middleware Order: Metis Research (Recovery → Logging → Security → Auth)
  - Security Headers: OWASP Guidelines
  - CORS: Explicit allowlist

  **Acceptance Criteria**:
  - [ ] Alle Middleware implementiert
  - [ ] Korrekte Reihenfolge: Recovery → RequestID → Logging → Security → CORS → Auth
  - [ ] JWT Auth Middleware setzt Claims in Context
  - [ ] API Key Auth Middleware validiert Scopes
  - [ ] Security Headers auf allen Responses
  - [ ] CORS mit expliziten Origins
  - [ ] Unit Tests für Middleware

  **QA Scenarios**:
  ```
  Scenario: Security Headers vorhanden
    Tool: Bash (curl)
    Steps:
      1. Starte API Server
      2. curl -I http://localhost:8080/health
    Expected Result: Headers enthalten X-Content-Type-Options, X-Frame-Options, CSP
    Evidence: curl Output mit Header-Liste
  ```

  **Commit**: YES
  - Message: `feat(security): implement auth and security middleware stack`
  - Files: `internal/middleware/*.go`

- [x] 1.8. Basis API Client (Mobile)

  **What to do**:
  - Erstelle API Client in `mobile/services/api.ts`:
    - Axios-Instance mit Basis-URL
    - Request Interceptor: Fügt JWT Token zu Header hinzu
    - Response Interceptor: Handled 401, refresht Token automatisch
    - Error Handling mit Retry-Logik
  - Erstelle API Service Funktionen (nur Interfaces, keine Implementierung):
    - `login(email, password)`
    - `register(email, password, displayName)`
    - `refreshToken(refreshToken)`
    - `logout()`
  - Erstelle Auth Store in `mobile/stores/authStore.ts`:
    - Zustand Store für Auth State
    - Methoden: login, logout, refresh, isAuthenticated
    - Persistenz mit AsyncStorage
  - Definiere TypeScript Interfaces für API Responses

  **Must NOT do**:
  - Keine UI implementieren
  - Keine echten API Calls (nur Client-Setup)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Mobile-Integration, Auth Flow
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T2.7 (Mobile Auth Flow)
  - **Blocked By**: T1.6

  **References**:
  - TanStack Query: Server State Management
  - Axios: HTTP Client mit Interceptors
  - Zustand: State Management

  **Acceptance Criteria**:
  - [ ] Axios-Instance mit Basis-URL konfiguriert
  - [ ] Request Interceptor fügt JWT Header hinzu
  - [ ] Response Interceptor handled 401 mit Token Refresh
  - [ ] Auth Store mit Zustand implementiert
  - [ ] Persistenz: Auth State in AsyncStorage
  - [ ] TypeScript Interfaces für API Responses

  **QA Scenarios**:
  ```
  Scenario: API Client ist konfiguriert
    Tool: TypeScript Compiler
    Steps:
      1. cd mobile && npx tsc --noEmit
    Expected Result: Keine TypeScript Fehler
    Evidence: Terminal Output "No errors"
  ```

  **Commit**: YES
  - Message: `feat(mobile): setup api client with axios and auth store`
  - Files: `mobile/services/api.ts`, `mobile/stores/authStore.ts`

- [x] 2.1. User Management API

  **What to do**:
  - Erstelle User Domain Model in `internal/domain/user.go`
  - Erstelle User Repository in `internal/repository/user_repo.go` mit pgx
  - Erstelle User Service in `internal/service/user_service.go` mit Business Logic
  - Erstelle User Handler in `internal/handler/user_handler.go`:
    - `POST /api/v1/auth/register` - Registrierung
    - `POST /api/v1/auth/login` - Login (gibt Token Pair)
    - `POST /api/v1/auth/refresh` - Token Refresh
    - `POST /api/v1/auth/logout` - Logout
    - `GET /api/v1/users/me` - Aktueller User
    - `PUT /api/v1/users/me` - Profil Update
  - Implementiere Password Hashing mit bcrypt
  - Validierung: Email-Format, Passwort-Stärke, Display Name
  - Tests für alle Endpoints

  **Must NOT do**:
  - Keine Push Notifications (kommt später)
  - Keine Gamification Logik

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: API Endpoints, CRUD Operationen
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: T2.2, T2.7
  - **Blocked By**: T1.4, T1.7

  **References**:
  - Clean Architecture: Handler → Service → Repository → Domain
  - Password Hashing: bcrypt mit Cost 12
  - Validation: `go-playground/validator`

  **Acceptance Criteria**:
  - [ ] User Domain Model definiert
  - [ ] User Repository mit pgx implementiert
  - [ ] User Service mit Business Logic
  - [ ] Alle Auth Endpoints funktionieren
  - [ ] Password Hashing mit bcrypt
  - [ ] Validierung: Email, Passwort (min 8 Zeichen), Display Name
  - [ ] Integration Tests mit Testcontainers

  **QA Scenarios**:
  ```
  Scenario: User Registration und Login
    Tool: Bash (curl)
    Steps:
      1. curl -X POST http://localhost:8080/api/v1/auth/register -d '{"email":"test@test.com","password":"secure123","display_name":"Test User"}'
      2. curl -X POST http://localhost:8080/api/v1/auth/login -d '{"email":"test@test.com","password":"secure123"}'
    Expected Result: Login gibt Access Token und Refresh Token zurück
    Evidence: curl Output mit Tokens
  ```

  **Commit**: YES
  - Message: `feat(api): implement user management and authentication endpoints`
  - Files: `internal/domain/user.go`, `internal/repository/user_repo.go`, `internal/service/user_service.go`, `internal/handler/user_handler.go`

- [x] 2.2. 1:1 Connections (Einladungen)

  **What to do**:
  - Erstelle Connection Domain Model in `internal/domain/connection.go`
  - Erstelle Connection Repository in `internal/repository/connection_repo.go`
  - Erstelle Connection Service in `internal/service/connection_service.go`:
    - Einladung generieren mit Token
    - Einladung akzeptieren/ablehnen
    - Verbindung auflösen
  - Erstelle Connection Handler in `internal/handler/connection_handler.go`:
    - `POST /api/v1/connections/invite` - Einladung generieren (gibt Link)
    - `GET /api/v1/connections/invite/:token` - Einladung validieren
    - `POST /api/v1/connections/invite/:token/accept` - Einladung annehmen
    - `POST /api/v1/connections/invite/:token/reject` - Einladung ablehnen
    - `GET /api/v1/connections` - Meine Verbindungen
    - `DELETE /api/v1/connections/:id` - Verbindung lösen
  - Einladungslink Format: `https://app.todo.com/invite/{token}`
  - Token: UUID, 24h gültig
  - Push Notification an Empfänger (Queue für später)

  **Must NOT do**:
  - Keine QR-Code Logik (kommt in T2.3)
  - Keine Deep Linking (Mobile kommt später)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Business Logic für Verbindungen
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: T2.3, T2.4, T4.3
  - **Blocked By**: T2.1

  **References**:
  - Connection Model: 1:1 Beziehung mit Status (pending, accepted, rejected)
  - Token: Kryptographisch sicher, zeitlich begrenzt

  **Acceptance Criteria**:
  - [ ] Connection Domain Model mit Status
  - [ ] Repository und Service implementiert
  - [ ] Alle Endpoints funktionieren
  - [ ] Einladungslink generiert mit 24h Ablauf
  - [ ] Nur 1 aktive Verbindung zwischen 2 Usern (unique constraint)
  - [ ] Integration Tests

  **QA Scenarios**:
  ```
  Scenario: Einladung erstellen und annehmen
    Tool: Bash (curl)
    Steps:
      1. User A: POST /connections/invite → Einladungslink
      2. User B: GET /connections/invite/{token} → Validiert
      3. User B: POST /connections/invite/{token}/accept
      4. User A: GET /connections → Zeigt Verbindung
    Expected Result: Beide User sind verbunden
    Evidence: API Responses
  ```

  **Commit**: YES
  - Message: `feat(api): implement 1:1 connection invitations`
  - Files: `internal/domain/connection.go`, `internal/repository/connection_repo.go`, `internal/service/connection_service.go`, `internal/handler/connection_handler.go`

- [x] 2.3. QR-Code Verbindung

  **What to do**:
  - Erweitere Connection Service:
    - `GenerateQRCode(userID)` → QR Code Payload
    - `ScanQRCode(scannerID, payload)` → Verbindung herstellen
  - QR Code Payload Format: JSON mit `{user_id, timestamp, signature}`
  - Signatur mit JWT Secret (verhindert Fälschung)
  - Zeitlich begrenzt: 5 Minuten gültig
  - Handler Endpoints:
    - `POST /api/v1/connections/qrcode/generate` - QR Code generieren
    - `POST /api/v1/connections/qrcode/scan` - QR Code scannen
  - Unit Tests für QR Code Logik

  **Must NOT do**:
  - Keine Mobile QR Scanner UI (kommt in Mobile Tasks)
  - Keine Bildgenerierung (nur Payload)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Kryptographische Signaturen
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: T2.7 (Mobile QR Scanner)
  - **Blocked By**: T2.2

  **References**:
  - QR Payload: JSON mit Signatur
  - Signatur: HMAC-SHA256 mit JWT Secret
  - Zeitfenster: 5 Minuten gegen Replay-Attacken

  **Acceptance Criteria**:
  - [ ] QR Code Generierung implementiert
  - [ ] Payload enthält user_id, timestamp, hmac signature
  - [ ] Scan validiert Signatur und Zeitfenster
  - [ ] Verbindung wird bei erfolgreichem Scan erstellt
  - [ ] Unit Tests für Signatur-Validierung

  **QA Scenarios**:
  ```
  Scenario: QR Code Generierung und Scan
    Tool: Bash (curl + Go Test)
    Steps:
      1. User A: POST /connections/qrcode/generate
      2. User B: POST /connections/qrcode/scan mit Payload
    Expected Result: Verbindung wird erstellt
    Evidence: API Response + DB Check
  ```

  **Commit**: YES
  - Message: `feat(api): add qr code connection flow`
  - Files: Erweiterung von Connection Service/Handler

- [x] 2.4. Todo CRUD API

  **What to do**:
  - Erstelle Todo Domain Model in `internal/domain/todo.go`:
    - Felder: id, title, description, status, priority, created_by, assigned_to, due_date, version (optimistic locking)
  - Erstelle Todo Repository in `internal/repository/todo_repo.go`
  - Erstelle Todo Service in `internal/service/todo_service.go`:
    - CRUD Operationen
    - Validierung: Title nicht leer, Status-Transitions
  - Erstelle Todo Handler in `internal/handler/todo_handler.go`:
    - `POST /api/v1/todos` - Todo erstellen
    - `GET /api/v1/todos` - Meine Todos (mit assigned_to oder created_by)
    - `GET /api/v1/todos/:id` - Todo Details
    - `PUT /api/v1/todos/:id` - Todo updaten
    - `DELETE /api/v1/todos/:id` - Todo löschen
  - Implementiere Optimistic Locking (Version Feld)
  - Tests für alle Endpoints

  **Must NOT do**:
  - Keine Gamification (kommt in Wave 3)
  - Keine Sync Logik (kommt in T2.5)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: CRUD API mit Optimistic Locking
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: T2.5, T2.8
  - **Blocked By**: T2.1

  **References**:
  - Optimistic Locking: Version Feld in DB, Prüfung bei Update
  - Status: pending → in_progress → completed

  **Acceptance Criteria**:
  - [ ] Todo Domain Model mit Version
  - [ ] CRUD Endpoints implementiert
  - [ ] Optimistic Locking: Version wird bei Update geprüft
  - [ ] Validierung: Title, Status-Transitions
  - [ ] GET /todos zeigt nur eigene Todos (created_by OR assigned_to)
  - [ ] Integration Tests

  **QA Scenarios**:
  ```
  Scenario: Todo CRUD Operations
    Tool: Bash (curl)
    Steps:
      1. POST /todos → Erstelle Todo
      2. GET /todos → Liste zeigt neues Todo
      3. PUT /todos/{id} → Update mit Version
      4. DELETE /todos/{id} → Löschen
    Expected Result: Alle Operationen erfolgreich
    Evidence: API Responses
  ```

  **Commit**: YES
  - Message: `feat(api): implement todo crud with optimistic locking`
  - Files: `internal/domain/todo.go`, `internal/repository/todo_repo.go`, `internal/service/todo_service.go`, `internal/handler/todo_handler.go`

- [x] 2.5. Sync Engine (Delta Sync)

  **What to do**:
  - Erstelle Sync Domain Model in `internal/domain/sync.go`:
    - ChangeSet: created[], updated[], deleted[]
    - Timestamps für last_pulled_at
  - Erstelle Sync Service in `internal/service/sync_service.go`:
    - `PullChanges(userID, lastPulledAt)` → Changes seit letztem Sync
    - `PushChanges(userID, changes)` → Client-Changes anwenden
  - Erstelle Sync Handler in `internal/handler/sync_handler.go`:
    - `POST /api/v1/sync` - Bidirektionaler Sync
  - Implementiere Delta Sync Logik:
    - Pull: SELECT * WHERE updated_at > last_pulled_at
    - Push: INSERT/UPDATE/DELETE mit Konflikt-Erkennung
  - Timestamps: Verwende Hybrid Logical Clocks (HLC)
  - Soft Delete: Markiere als deleted statt hard delete

  **Must NOT do**:
  - Keine Konfliktlösung (kommt in T2.6)
  - Keine Offline-First (Mobile kommt später)

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Komplexe Sync-Logik, Timestamps, Delta-Algorithmus
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: T2.6, T4.4
  - **Blocked By**: T2.4

  **References**:
  - Delta Sync: Metis Research (Pull/Push mit Timestamps)
  - HLC: Hybrid Logical Clocks für verteilte Systeme
  - Soft Delete: Tombstones statt DELETE

  **Acceptance Criteria**:
  - [ ] Sync Service mit Pull/Push implementiert
  - [ ] Delta Sync: Nur Änderungen seit last_pulled_at
  - [ ] Timestamps mit HLC
  - [ ] Soft Delete: deleted_at statt DELETE
  - [ ] Konflikt-Erkennung (beide Seiten geändert)
  - [ ] Integration Tests mit Sync-Szenarien

  **QA Scenarios**:
  ```
  Scenario: Delta Sync zwischen Client und Server
    Tool: Bash (curl)
    Steps:
      1. Client: POST /sync mit changes → Push
      2. Server: Verarbeitet Änderungen
      3. Client: POST /sync mit last_pulled_at → Pull
    Expected Result: Client erhält server-seitige Änderungen
    Evidence: API Response mit changes
  ```

  **Commit**: YES
  - Message: `feat(api): implement delta sync engine`
  - Files: `internal/domain/sync.go`, `internal/service/sync_service.go`, `internal/handler/sync_handler.go`

- [x] 2.6. Konfliktlösung

  **What to do**:
  - Erstelle Conflict Resolution Service in `internal/service/conflict_service.go`:
    - `DetectConflicts(userID, changes)` → Konflikte identifizieren
    - `ResolveConflict(table, local, remote, strategy)` → Konflikt lösen
  - Implementiere Konfliktlösungs-Strategien pro Feld:
    - Title/Description: "prompt" → Markiere für User-Resolution
    - Status: "last-write-wins" → Höherer Timestamp gewinnt
    - Due Date: "merge" → Späteres Datum gewinnt
    - Priority: "max-wins" → Höhere Priorität gewinnt
  - Erweitere Sync Handler:
    - Gib Konflikte in Response zurück: `{timestamp, changes, conflicts}`
  - Erstelle sync_conflicts Tabelle:
    - Speichert Konflikte für User-Resolution
    - Fields: user_id, table_name, record_id, local_data, server_data, status
  - Unit Tests für alle Resolution-Strategien

  **Must NOT do**:
  - Keine UI für Konflikt-Resolution (Mobile kommt später)

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Komplexe Konfliktlösungs-Logik
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: T4.4 (Mobile Offline)
  - **Blocked By**: T2.5

  **References**:
  - Conflict Resolution: Metis Research (Per-Feld Strategien)
  - WatermelonDB: Custom Resolver verhindert Datenverlust

  **Acceptance Criteria**:
  - [ ] Conflict Service implementiert
  - [ ] Strategien pro Feld-Typ definiert
  - [ ] Konflikte werden in Response zurückgegeben
  - [ ] sync_conflicts Tabelle speichert ungelöste Konflikte
  - [ ] Unit Tests für alle Resolution-Strategien

  **QA Scenarios**:
  ```
  Scenario: Konflikt-Erkennung und Lösung
    Tool: Bash (curl + Go Test)
    Steps:
      1. Server-Seite ändert Todo Title
      2. Client pusht andere Änderung an selbem Todo
      3. Sync Response enthält Konflikt
    Expected Result: Konflikt wird erkannt und zurückgegeben
    Evidence: API Response mit conflicts Array
  ```

  **Commit**: YES
  - Message: `feat(api): implement conflict detection and resolution`
  - Files: `internal/service/conflict_service.go`, Migrations für sync_conflicts

- [x] 2.7. Mobile Auth Flow

  **What to do**:
  - Erstelle Auth Screens in `mobile/app/`:
    - `login.tsx` - Login Form
    - `register.tsx` - Registrierung Form
    - `_layout.tsx` - Auth Stack Navigator
  - Implementiere Login/Register mit API Client:
    - Form Validation (Email, Passwort)
    - API Calls: login, register
    - Token Speicherung in Auth Store
  - Erstelle Protected Route Wrapper:
    - Prüft Auth State
    - Redirect zu Login wenn nicht authentifiziert
  - Erstelle `app/(app)/_layout.tsx` - Main App Layout (nach Auth)
  - Erstelle `app/(app)/index.tsx` - Home Screen (Dashboard)
  - Implementiere Logout Funktion
  - Tests für Auth Flow

  **Must NOT do**:
  - Keine Todo UI (kommt in T2.8)
  - Keine Verbindungen UI (kommt später)

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Mobile UI, Forms, Navigation
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: T2.8, T3.6, T3.8
  - **Blocked By**: T1.6, T1.8, T2.1

  **References**:
  - Expo Router: File-based Navigation
  - React Hook Form: Form Management
  - Zustand: Auth State

  **Acceptance Criteria**:
  - [ ] Login und Register Screens implementiert
  - [ ] Form Validation funktioniert
  - [ ] API Calls für Auth
  - [ ] Tokens werden gespeichert und verwendet
  - [ ] Protected Routes funktionieren
  - [ ] Logout funktioniert
  - [ ] E2E Test mit Maestro

  **QA Scenarios**:
  ```
  Scenario: User kann sich registrieren und einloggen
    Tool: Maestro (E2E)
    Steps:
      1. Starte App
      2. Tippe "Register"
      3. Fülle Formular aus
      4. Tippe "Submit"
      5. Überprüfe: Redirect zu Home Screen
    Expected Result: User ist eingeloggt, sieht Dashboard
    Evidence: Maestro Screenshot
  ```

  **Commit**: YES
  - Message: `feat(mobile): implement authentication screens and flow`
  - Files: `mobile/app/login.tsx`, `mobile/app/register.tsx`, `mobile/app/(app)/**`

- [x] 2.8. Mobile Todo UI

  **What to do**:
  - Erstelle Todo List Screen in `mobile/app/(app)/todos.tsx`:
    - Liste aller Todos
    - Pull-to-Refresh für Sync
    - Filter: Alle, Offen, Erledigt
  - Erstelle Todo Detail Screen in `mobile/app/(app)/todos/[id].tsx`:
    - Todo Details anzeigen
    - Status ändern
    - Löschen
  - Erstelle Create/Edit Todo Screen:
    - Form für Title, Description, Due Date, Priority
    - Zuweisung an Connection Partner
  - Implementiere Todo Store mit Zustand:
    - Todos State
    - CRUD Actions
  - Verwende `@shopify/flash-list` für Performance
  - Tests für Todo UI

  **Must NOT do**:
  - Keine Offline-First (kommt in T4.4)
  - Keine Sync-Integration (kommt später)

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Mobile UI, Listen, Forms
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: T3.6, T3.8, T4.4
  - **Blocked By**: T2.4, T2.7

  **References**:
  - Flash List: Performance für große Listen
  - React Native Paper oder NativeWind: UI Components

  **Acceptance Criteria**:
  - [ ] Todo List Screen implementiert
  - [ ] Todo Detail Screen implementiert
  - [ ] Create/Edit Todo Screen implementiert
  - [ ] CRUD Operationen funktionieren
  - [ ] Pull-to-Refresh implementiert
  - [ ] Filter funktionieren
  - [ ] E2E Tests mit Maestro

  **QA Scenarios**:
  ```
  Scenario: User kann Todo erstellen und abschließen
    Tool: Maestro (E2E)
    Steps:
      1. Login
      2. Tippe "+" für neues Todo
      3. Fülle Title ein
      4. Tippe "Speichern"
      5. Tippe auf Todo in Liste
      6. Tippe "Abschließen"
    Expected Result: Todo ist erledigt, Status aktualisiert
    Evidence: Maestro Screenshot
  ```

  **Commit**: YES
  - Message: `feat(mobile): implement todo list and detail screens`
  - Files: `mobile/app/(app)/todos.tsx`, `mobile/app/(app)/todos/[id].tsx`, etc.

- [x] 3.1. Gamification Domain Model

  **What to do**:
  - Erstelle Domain Models in `internal/domain/gamification.go`:
    - `UserStats` - Punkte, Level, Streak, Badges
    - `Badge` - Badge Definitionen
    - `Level` - Level Definitionen mit XP requirements
    - `PointsTransaction` - Immutable Ledger für Punkte
  - Definiere Level-Kurve:
    - Level 1: 0 XP
    - Level 2: 100 XP
    - Level 3: 300 XP
    - Level 4: 700 XP
    - Level 5: 1500 XP
    - Level 6: 3000 XP
    - Level 7: 6000 XP
    - Level 8: 12000 XP
  - Definiere XP Rewards:
    - Todo Completed: 10 XP
    - Streak Bonus 7 Days: 50 XP
    - Streak Bonus 30 Days: 200 XP
    - Perfect Day (alle Todos): 25 XP
    - Badge Earned: Je nach Badge
  - Erstelle Formel: `CalculateLevel(xp int) int`
  - Unit Tests für Level-Berechnung

  **Must NOT do**:
  - Keine Repository/Service Logik (nur Domain Models)
  - Keine Badges definieren (kommt in T3.3)

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Business Logic für Gamification
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T3.2, T3.3, T3.4, T3.5
  - **Blocked By**: T2.4

  **References**:
  - Level Curve: Metis Research (lineare Progression)
  - XP Rewards: Metis Research (sinnvolle Belohnungen)

  **Acceptance Criteria**:
  - [ ] Alle Domain Models definiert
  - [ ] Level-Kurve implementiert
  - [ ] XP Rewards definiert
  - [ ] `CalculateLevel` funktioniert korrekt
  - [ ] Unit Tests für alle Berechnungen

  **QA Scenarios**:
  ```
  Scenario: Level Berechnung
    Tool: Bash (Go Test)
    Steps:
      1. go test -v ./internal/domain/... -run TestGamification
    Expected Result: Alle Level-Berechnungen korrekt
    Evidence: Test Output
  ```

  **Commit**: YES
  - Message: `feat(gamification): define domain models and level curve`
  - Files: `internal/domain/gamification.go`, `internal/domain/gamification_test.go`

- [x] 3.2. XP & Level System

  **What to do**:
  - Erstelle Gamification Repository in `internal/repository/gamification_repo.go`
  - Erstelle Gamification Service in `internal/service/gamification_service.go`:
    - `AwardXP(userID, amount, actionType, reference)` → Punkte vergeben
    - `CalculateLevel(xp)` → Level berechnen
    - `UpdateStreak(userID)` → Streak aktualisieren
    - `GetUserStats(userID)` → Stats abrufen
  - Implementiere Streak-Logik:
    - Daily Login → Streak +1
    - Verpasster Tag → Streak reset (aber: 1 Freeze Token pro Monat)
  - Erstelle Handler Endpoints:
    - `GET /api/v1/users/me/stats` → User Stats
    - `GET /api/v1/users/me/history` → Punkte-Historie
  - Integration mit Todo Completion:
    - Wenn Todo completed → XP vergeben
    - Wenn Streak Milestone → Bonus XP
  - Tests für alle Aktionen

  **Must NOT do**:
  - Keine Badges (kommt in T3.3)
  - Keine Anti-Cheat (kommt in T3.4)

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Komplexe Business Logic (Streaks, XP, Level-Ups)
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T3.3, T3.6
  - **Blocked By**: T3.1

  **References**:
  - Streak Logic: Metis Research (Freeze Tokens für Flexibilität)
  - XP Ledger: Immutable, nie löschen oder ändern

  **Acceptance Criteria**:
  - [ ] Repository und Service implementiert
  - [ ] XP wird bei Todo Completion vergeben
  - [ ] Level-Up wird berechnet und gespeichert
  - [ ] Streak-Logik funktioniert (inkl. Freeze Tokens)
  - [ ] Punkte-Historie wird geführt
  - [ ] Integration Tests

  **QA Scenarios**:
  ```
  Scenario: XP Vergabe und Level-Up
    Tool: Bash (curl + Go Test)
    Steps:
      1. User completed Todo → 10 XP
      2. Prüfe Stats: XP +10
      3. Wiederhole bis Level-Up
      4. Prüfe: Neues Level
    Expected Result: XP korrekt, Level-Up funktioniert
    Evidence: API Response + DB Check
  ```

  **Commit**: YES
  - Message: `feat(gamification): implement xp and level system`
  - Files: `internal/repository/gamification_repo.go`, `internal/service/gamification_service.go`

- [x] 3.3. Badge System

  **What to do**:
  - Definiere Badges in Datenbank (Seed Data):
    - "First Todo" (1 Todo completed) - 10 XP
    - "Week Warrior" (7 Tage Streak) - 50 XP
    - "Century Club" (100 Todos) - 100 XP
    - "Connection Master" (10 Verbindungen) - 75 XP
    - "Early Bird" (Erster Todo am Tag) - 25 XP
  - Erweitere Gamification Service:
    - `CheckAndAwardBadges(userID)` → Prüft alle Badges
    - `EvaluateBadgeCriteria(userID, badge)` → Prüft spezifisches Badge
  - Implementiere Badge-Kriterien:
    - Todos Completed Count
    - Streak Days
    - Connections Count
    - Early Completion
  - Erstelle Handler:
    - `GET /api/v1/users/me/badges` → Meine Badges
  - Badge wird automatisch vergeben wenn Kriterien erfüllt
  - Tests für Badge-Evaluation

  **Must NOT do**:
  - Keine "Secret Badges" (optional für später)
  - Keine Badge-UI (kommt in T3.6)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Badge-Evaluation Logik
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T3.6
  - **Blocked By**: T3.2

  **References**:
  - Badge Patterns: Metis Research (sinnvolle Kriterien)
  - Automatische Vergabe bei Kriterien-Erfüllung

  **Acceptance Criteria**:
  - [ ] Alle Badges definiert (Seed Data)
  - [ ] Badge-Evaluation implementiert
  - [ ] Badges werden automatisch vergeben
  - [ ] XP Bonus bei Badge-Erhalt
  - [ ] Handler für Badge-Liste
  - [ ] Integration Tests

  **QA Scenarios**:
  ```
  Scenario: Badge wird automatisch vergeben
    Tool: Bash (curl)
    Steps:
      1. User completed 1. Todo
      2. System prüft Badges
      3. Badge "First Todo" wird vergeben
      4. GET /users/me/badges zeigt Badge
    Expected Result: Badge in Liste, XP Bonus erhalten
    Evidence: API Response
  ```

  **Commit**: YES
  - Message: `feat(gamification): implement badge system with auto-award`
  - Files: Erweiterung Gamification Service, Seed Data

- [x] 3.4. Anti-Cheat Validation

  **What to do**:
  - Erstelle AntiCheat Service in `internal/service/anticheat_service.go`:
    - `ValidateTodoComplete(userID, todoID)` → Prüft ob Aktion gültig
    - Rate Limiting: Max 10 Completions pro Minute
    - Timestamp Validation: Client-Zeit ±5 Min von Server
    - Idempotenz: Keine doppelten Submits (24h TTL)
    - Business Rules: Todo nicht bereits completed
  - Implementiere Validation Rules:
    - Minimale Zeit zwischen Actions: 5 Sekunden
    - Kein Backdating (completed_at muss >= created_at sein)
    - Keine Rapid Complete/Uncomplete Cycling
  - Verwende Redis für:
    - Rate Limiting Counters
    - Idempotenz Keys
    - Action Timestamps
  - Erweitere Todo Completion Flow:
    - Vor XP-Vergabe: AntiCheat prüfen
    - Bei Verdacht: Aktion ablehnen, loggen
  - Tests für alle Anti-Cheat Rules

  **Must NOT do**:
  - Keine User-Sperrung (nur Aktion ablehnen)
  - Keine komplexe ML-Detection (für MVP zu aufwändig)

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Security-kritisch, muss Exploits verhindern
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T3.2 (Modifikation)
  - **Blocked By**: T1.1 (Redis muss laufen)

  **References**:
  - Anti-Cheat: Metis Research (Server-side Validation)
  - Redis: Rate Limiting, Idempotenz

  **Acceptance Criteria**:
  - [ ] AntiCheat Service implementiert
  - [ ] Rate Limiting: 10 Actions/Minute
  - [ ] Idempotenz: 24h TTL
  - [ ] Timestamp Validation ±5 Min
  - [ ] Business Rules geprüft
  - [ ] Bei Exploit-Versuch: Aktion ablehnen + loggen
  - [ ] Unit Tests für Anti-Cheat

  **QA Scenarios**:
  ```
  Scenario: Anti-Cheat blockiert Exploit
    Tool: Bash (curl)
    Steps:
      1. User versucht Todo 2x schnell zu completen
      2. 2. Request sollte blockiert werden
    Expected Result: 2. Request gibt 429 Too Many Requests
    Evidence: API Response mit Rate Limit Header
  ```

  **Commit**: YES
  - Message: `feat(security): implement anti-cheat validation system`
  - Files: `internal/service/anticheat_service.go`, `internal/service/anticheat_service_test.go`

- [x] 3.5. Custom Rewards

  **What to do**:
  - Erstelle Reward Domain Model in `internal/domain/reward.go`:
    - `Reward` - Belohnungs-Definition (Name, Beschreibung, Kosten)
    - `RewardRedemption` - Einlösung (Status, Zeitpunkt)
  - Erstelle Reward Repository in `internal/repository/reward_repo.go`
  - Erstelle Reward Service in `internal/service/reward_service.go`:
    - `CreateReward(userID, name, description, cost)` → Belohnung erstellen
    - `RedeemReward(userID, rewardID)` → Belohnung einlösen
    - `ApproveRedemption(redemptionID)` → Admin Approval
    - `ListRewards(userID)` → Verfügbare Belohnungen
    - `ListMyRewards(userID)` → Meine Einlösungen
  - Erstelle Handler:
    - `POST /api/v1/rewards` - Belohnung erstellen
    - `GET /api/v1/rewards` - Liste aller Belohnungen
    - `POST /api/v1/rewards/:id/redeem` - Einlösen
    - `GET /api/v1/rewards/my` - Meine Einlösungen
  - Logik:
    - Erstellen: Jeder User kann Belohnungen definieren
    - Einlösen: Kosten werden von XP abgezogen
    - Approval: Optional, je nach Konfiguration
  - Tests für Reward Flow

  **Must NOT do**:
  - Keine Echtzeit-Benachrichtigung bei Einlösung (kommt in T4.1)
  - Keine Bild-Uploads für Belohnungen

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: CRUD API mit Business Logic
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T3.6
  - **Blocked By**: T3.2

  **References**:
  - Reward Pattern: Metis Research (User-definierte Belohnungen)
  - Einlösung: XP-Kosten, Admin-Approval optional

  **Acceptance Criteria**:
  - [ ] Repository und Service implementiert
  - [ ] Belohnungen erstellen funktioniert
  - [ ] Einlösung zieht XP ab
  - [ ] Liste aller Belohnungen
  - [ ] Meine Einlösungen tracken
  - [ ] Integration Tests

  **QA Scenarios**:
  ```
  Scenario: Belohnung erstellen und einlösen
    Tool: Bash (curl)
    Steps:
      1. POST /rewards → Erstelle Belohnung (100 XP)
      2. POST /rewards/{id}/redeem → Einlösen
      3. Prüfe: XP wurden abgezogen
      4. GET /rewards/my → Zeigt Einlösung
    Expected Result: Belohnung erstellt, XP abgezogen, Einlösung gespeichert
    Evidence: API Responses + DB Check
  ```

  **Commit**: YES
  - Message: `feat(rewards): implement custom reward system`
  - Files: `internal/domain/reward.go`, `internal/repository/reward_repo.go`, `internal/service/reward_service.go`

- [x] 3.6. Gamification UI

  **What to do**:
  - Erstelle Stats Screen in `mobile/app/(app)/stats.tsx`:
    - Punkte-Anzeige
    - Level mit Fortschrittsbalken
    - Aktuelle Streak
    - Badges Grid
  - Erstelle Rewards Screen in `mobile/app/(app)/rewards.tsx`:
    - Verfügbare Belohnungen
    - Meine Einlösungen
    - Belohnung erstellen Form
  - Erstelle Components:
    - `ProgressBar` - Für Level-Fortschritt
    - `BadgeCard` - Badge Anzeige
    - `RewardCard` - Belohnungs-Karte
  - Verwende TanStack Query für Data Fetching
  - Animationen mit Reanimated:
    - Level-Up Animation
    - XP Gain Animation
    - Badge Earned Celebration
  - Tests für UI

  **Must NOT do**:
  - Keine komplexen Charts (für MVP zu aufwändig)
  - Keine 3D Animationen

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Mobile UI, Animationen
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T3.8
  - **Blocked By**: T3.2, T3.3, T3.5, T2.7

  **References**:
  - Reanimated 4: Animations
  - Lottie: Complex Celebrations
  - TanStack Query: Data Fetching

  **Acceptance Criteria**:
  - [ ] Stats Screen zeigt Punkte, Level, Streak, Badges
  - [ ] Level-Up Animation
  - [ ] Rewards Screen mit Liste
  - [ ] Belohnung erstellen funktioniert
  - [ ] XP Gain Animation bei Todo Completion
  - [ ] E2E Tests

  **QA Scenarios**:
  ```
  Scenario: User sieht seine Stats
    Tool: Maestro (E2E)
    Steps:
      1. Login
      2. Navigiere zu "Stats"
      3. Prüfe: Punkte, Level, Badges werden angezeigt
    Expected Result: Alle Stats sichtbar
    Evidence: Maestro Screenshot
  ```

  **Commit**: YES
  - Message: `feat(mobile): implement gamification ui with animations`
  - Files: `mobile/app/(app)/stats.tsx`, `mobile/app/(app)/rewards.tsx`, `mobile/components/ProgressBar.tsx`, etc.

- [x] 3.7. Collaborative Goals

  **What to do**:
  - Erstelle SharedGoal Domain Model in `internal/domain/shared_goal.go`:
    - `SharedGoal` - Ziel für 2 verbundene User
    - Felder: connection_id, target_type, target_value, current_value, reward_description
  - Erstelle Repository und Service:
    - `CreateGoal(connectionID, targetType, targetValue, reward)`
    - `UpdateProgress(connectionID, amount)` → Fortschritt aktualisieren
    - `CheckCompletion(goalID)` → Prüft ob Ziel erreicht
    - `ListGoals(userID)` → Meine Shared Goals
  - Ziel-Typen:
    - `todos_completed` - Gemeinsam X Todos erledigen
    - `streak_days` - X Tage Streak beide User
  - Automatische Progress-Updates:
    - Wenn Todo completed → Beide Todos zählen für Goal
    - Wenn Goal erreicht → Push Notification + XP Bonus
  - Erstelle Handler:
    - `POST /api/v1/goals` - Ziel erstellen
    - `GET /api/v1/goals` - Meine Ziele
  - Tests

  **Must NOT do**:
  - Keine komplexen Ziel-Typen (nur 2 einfache)
  - Keine Wettkämpfe (nur kooperativ)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Business Logic für Shared Goals
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T4.1 (Push Notifications)
  - **Blocked By**: T2.2 (Connections), T3.2

  **References**:
  - Collaborative Goals: Metis Research (statt Leaderboards)
  - Progress: Beide User tragen bei

  **Acceptance Criteria**:
  - [ ] SharedGoal Domain Model
  - [ ] Repository und Service implementiert
  - [ ] 2 Ziel-Typen: todos_completed, streak_days
  - [ ] Progress wird automatisch aktualisiert
  - [ ] Bei Completion: XP Bonus
  - [ ] Integration Tests

  **QA Scenarios**:
  ```
  Scenario: Shared Goal wird erreicht
    Tool: Bash (curl)
    Steps:
      1. User A und B sind verbunden
      2. Erstelle Goal: 10 Todos gemeinsam
      3. Beide User erledigen Todos
      4. Progress steigt
      5. Bei 10/10: Goal completed
    Expected Result: Progress korrekt, Bonus XP
    Evidence: API Responses
  ```

  **Commit**: YES
  - Message: `feat(goals): implement collaborative shared goals`
  - Files: `internal/domain/shared_goal.go`, Repository, Service, Handler

- [x] 3.8. Animationen & Celebration

  **What to do**:
  - Erstelle Animation Components:
    - `LevelUpAnimation` - Confetti + Level-Anzeige
    - `XPGainPopup` - Floating XP Text
    - `BadgeEarnedModal` - Badge mit Celebration
    - `GoalCompletedAnimation` - Shared Goal Erfolg
  - Verwende:
    - Reanimated 4 für smooth animations
    - Lottie für komplexe Celebration-Animationen
  - Trigger Points:
    - Todo Completed → XP Popup
    - Level Up → LevelUpAnimation
    - Badge Earned → BadgeEarnedModal
    - Goal Completed → GoalCompletedAnimation
  - Integration in bestehende Screens
  - Tests

  **Must NOT do**:
  - Keine übermäßigen Animationen (nicht bei jedem Action)
  - Keine 3D oder WebGL

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Animationen, UI Polish
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T4.x
  - **Blocked By**: T3.6

  **References**:
  - Reanimated 4: Worklet-based animations
  - Lottie: After Effects Animationen

  **Acceptance Criteria**:
  - [ ] Alle Animation Components implementiert
  - [ ] Level-Up Animation bei Level-Up
  - [ ] XP Popup bei Todo Completion
  - [ ] Badge Modal bei Badge-Erhalt
  - [ ] Goal Animation bei Goal Completion
  - [ ] Flüssige 60fps Animationen

  **QA Scenarios**:
  ```
  Scenario: Level-Up Animation wird angezeigt
    Tool: Maestro (E2E)
    Steps:
      1. User completed genug Todos für Level-Up
      2. Animation wird angezeigt
      3. Neues Level wird gezeigt
    Expected Result: Animation flüssig, Level korrekt
    Evidence: Maestro Screenshot + Video
  ```

  **Commit**: YES
  - Message: `feat(ui): add celebration animations and gamification feedback`
  - Files: `mobile/components/animations/*.tsx`

- [x] 4.1. Push Notification Queue

  **What to do**:
  - Erstelle NotificationQueue in PostgreSQL (Tabelle bereits in T1.2 erstellt)
  - Erstelle Notification Service in `internal/service/notification_service.go`:
    - `QueueNotification(userID, title, body, data, priority)` → Fügt zu Queue hinzu
    - `ProcessQueue()` → Verarbeitet Pending Notifications
  - Trigger Points:
    - Connection Request → Notification an Empfänger
    - Todo Assigned → Notification an Assignee
    - Todo Completed → Notification an Creator
    - Badge Earned → Notification an User
    - Goal Completed → Notification an beide Partner
  - Queue Features:
    - Priorisierung (1-5)
    - Retry Logik (max 3 Versuche)
    - Scheduled At (für verzögerte Zustellung)
    - Error Tracking
  - Tests

  **Must NOT do**:
  - Kein echter Push Versand (kommt in T4.2)
  - Keine WebSocket Realtime (zu komplex für MVP)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Queue-Management, Retry-Logik
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: T4.2
  - **Blocked By**: T2.2, T2.4, T3.3, T3.7

  **References**:
  - Queue Pattern: Metis Research (PostgreSQL Queue)
  - Retry: Exponentielles Backoff

  **Acceptance Criteria**:
  - [ ] Notification Service implementiert
  - [ ] Alle Trigger Points erstellt
  - [ ] Queue Einträge mit Priority
  - [ ] Retry Logik (max 3)
  - [ ] Error Tracking
  - [ ] Integration Tests

  **QA Scenarios**:
  ```
  Scenario: Notification wird in Queue eingefügt
    Tool: Bash (curl + SQL)
    Steps:
      1. User A sendet Connection Request an User B
      2. Prüfe DB: notification_queue hat Eintrag für User B
    Expected Result: Queue Eintrag mit Status 'pending'
    Evidence: SQL Query Result
  ```

  **Commit**: YES
  - Message: `feat(notifications): implement notification queue system`
  - Files: `internal/service/notification_service.go`

- [x] 4.2. Notification Worker

  **What to do**:
  - Erstelle Worker in `cmd/worker/main.go`:
    - Separater Prozess vom API Server
    - Pollt notification_queue alle 10 Sekunden
    - Verarbeitet Pending Notifications
  - Implementiere Push Versand:
    - FCM für Android (Firebase Admin SDK)
    - APNS für iOS (expo-server-sdk oder sideshow/apns2)
    - Expo Push API als Alternative
  - Verarbeitungs-Flow:
    1. SELECT pending notifications
    2. Für jede: Sende Push
    3. Bei Erfolg: Mark as sent
    4. Bei Fehler: Retry counter +1, schedule retry
    5. Nach 3 Fehlern: Dead letter queue
  - Token Management:
    - Hole Device Tokens aus DB
    - Bei "Invalid Token": Entferne aus DB
  - Graceful Shutdown für Worker
  - Tests

  **Must NOT do**:
  - Keine In-App Notifications (nur Push)
  - Keine Email Notifications

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Background Worker, Push APIs
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: None
  - **Blocked By**: T4.1

  **References**:
  - FCM: Firebase Admin SDK Go
  - APNS: expo-server-sdk oder sideshow/apns2
  - Worker Pattern: Separate Prozess, Polling

  **Acceptance Criteria**:
  - [ ] Worker Prozess implementiert
  - [ ] FCM Integration für Android
  - [ ] APNS Integration für iOS
  - [ ] Retry Logik mit Backoff
  - [ ] Dead Letter Queue nach 3 Fehlern
  - [ ] Invalid Token Cleanup
  - [ ] Graceful Shutdown
  - [ ] Tests

  **QA Scenarios**:
  ```
  Scenario: Notification wird zugestellt
    Tool: Bash + Expo Push Receipt
    Steps:
      1. Worker läuft
      2. Queue hat Pending Notification
      3. Worker verarbeitet Queue
      4. Prüfe Expo Push Receipt
    Expected Result: Push Receipt zeigt "ok"
    Evidence: Expo Push API Response
  ```

  **Commit**: YES
  - Message: `feat(worker): implement push notification worker with fcm/apns`
  - Files: `cmd/worker/main.go`, Push Provider Implementierungen

- [x] 4.3. Deep Linking (Einladungen)

  **What to do**:
  - Konfiguriere Deep Linking in `mobile/app.json`:
    - Scheme: `todoapp://`
    - Domain: `app.deinedomain.com`
  - Erstelle `mobile/app/invite.tsx`:
    - Deep Link Handler für `/invite/{token}`
    - Zeigt Einladungs-Details
    - Buttons: Annehmen / Ablehnen
  - Implementiere Universal Links (iOS) und App Links (Android)
  - Erweitere Connection Service:
    - Deep Link generiert automatisch Einladungslink
  - Integration:
    - App öffnet sich bei Klick auf Einladungslink
    - Nicht-authentifizierte User → Login → dann Invite Screen
  - Tests

  **Must NOT do**:
  - Keine QR-Code Deep Links (QR wird in App gescannt)
  - Keine Fallback Webseite

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Mobile Integration, Deep Linking
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: None
  - **Blocked By**: T2.2, T2.7

  **References**:
  - Expo Linking: `expo-linking`
  - Universal Links: iOS
  - App Links: Android

  **Acceptance Criteria**:
  - [ ] Deep Linking konfiguriert
  - [ ] Invite Screen implementiert
  - [ ] Universal Links funktionieren
  - [ ] App Links funktionieren
  - [ ] Unauthentifizierte User werden zu Login weitergeleitet
  - [ ] E2E Test

  **QA Scenarios**:
  ```
  Scenario: User öffnet Einladungslink
    Tool: Maestro (E2E)
    Steps:
      1. Öffne Link: todoapp://invite/abc123
      2. App öffnet Invite Screen
      3. Zeigt Details zur Einladung
      4. User tippt "Annehmen"
    Expected Result: Verbindung wird hergestellt
    Evidence: Maestro Screenshot
  ```

  **Commit**: YES
  - Message: `feat(mobile): implement deep linking for invitations`
  - Files: `mobile/app/invite.tsx`, `mobile/app.json` Konfiguration

- [x] 4.4. Offline-First Sync

  **What to do**:
  - Integriere WatermelonDB in Mobile App:
    - Setup Database in `mobile/services/database.ts`
    - Definiere Schemas: Users, Todos, Connections, etc.
  - Implementiere Sync mit Backend:
    - `sync()` Funktion mit TanStack Query
    - Pull: Hole Änderungen vom Server
    - Push: Sende lokale Änderungen
    - Konfliktlösung: Zeige Konflikte UI
  - Optimistic Updates:
    - Todo completed → Sofort UI Update
    - Dann Sync im Hintergrund
    - Bei Fehler: Rollback mit Toast
  - Offline Indicators:
    - Connection Status Badge
    - Pending Changes Counter
    - "Syncing..." Spinner
  - Background Sync:
    - Bei App-Foreground
    - Bei Pull-to-Refresh
    - Periodisch (30s wenn aktiv)
  - Tests

  **Must NOT do**:
  - Keine WebSocket Realtime (nur Polling)
  - Keine komplexe CRDT Logic

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Komplexe Sync-Logik, Offline-First
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: None
  - **Blocked By**: T2.5, T2.6, T2.8

  **References**:
  - WatermelonDB: `@nozbe/watermelondb`
  - Sync Protocol: Pull/Push mit Timestamps
  - Optimistic UI: TanStack Query

  **Acceptance Criteria**:
  - [ ] WatermelonDB integriert
  - [ ] Sync funktioniert: Pull + Push
  - [ ] Optimistic Updates
  - [ ] Offline Indicators
  - [ ] Background Sync
  - [ ] Konflikte werden angezeigt
  - [ ] E2E Tests

  **QA Scenarios**:
  ```
  Scenario: Offline Todo wird synchronisiert
    Tool: Maestro (E2E)
    Steps:
      1. Erstelle Todo offline
      2. Verbinde mit Internet
      3. Sync wird ausgeführt
      4. Prüfe Backend: Todo existiert
    Expected Result: Todo wurde synchronisiert
    Evidence: API Response + DB Check
  ```

  **Commit**: YES
  - Message: `feat(mobile): implement offline-first sync with watermelondb`
  - Files: `mobile/services/database.ts`, `mobile/services/sync.ts`, Sync Integration in UI

- [x] 4.5. Observability & Monitoring

  **What to do**:
  - Implementiere Structured Logging:
    - `rs/zerolog` für JSON Logs
    - Request ID in jedem Log
    - User ID, Endpoint, Duration, Status
  - Erstelle Metriken mit Prometheus:
    - `http_request_duration_seconds` - Histogram
    - `todos_completed_total` - Counter
    - `sync_duration_seconds` - Histogram
    - `active_users` - Gauge
  - Erstelle Health Endpoints:
    - `GET /health/live` - Liveness (immer 200)
    - `GET /health/ready` - Readiness (prüft DB, Redis)
  - Implementiere Tracing (optional aber empfohlen):
    - OpenTelemetry für verteiltes Tracing
    - Trace IDs über Request Chain
  - Error Tracking:
    - Sentry Integration (optional)
    - Error Reports mit Context
  - Dashboards:
    - Grafana für Metriken
    - Loki für Logs (optional)
  - Tests

  **Must NOT do**:
  - Keine komplexe APM Tools (New Relic, etc.)
  - Keine User-Tracking (nur technische Metriken)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Setup und Integration
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: T4.8
  - **Blocked By**: T1.3

  **References**:
  - Structured Logging: zerolog
  - Prometheus: Metrics
  - Health Checks: Kubernetes Best Practices

  **Acceptance Criteria**:
  - [ ] Structured Logging implementiert
  - [ ] Prometheus Metrics exponiert
  - [ ] Health Endpoints funktionieren
  - [ ] Request Tracing (optional)
  - [ ] Error Tracking (optional)
  - [ ] Tests

  **QA Scenarios**:
  ```
  Scenario: Health Check gibt Status
    Tool: Bash (curl)
    Steps:
      1. curl http://localhost:8080/health/ready
    Expected Result: JSON mit Status und Checks
    Evidence: curl Output
  ```

  **Commit**: YES
  - Message: `feat(observability): add logging, metrics, and health checks`
  - Files: `internal/observability/`, Health Handler

- [x] 4.6. CI/CD Pipeline

  **What to do**:
  - Backend CI/CD:
    - GitHub Actions Workflow: `.github/workflows/backend.yml`
    - Lint: `golangci-lint`
    - Test: `go test` mit Testcontainers
    - Build: Docker Image bauen
    - Security: `gosec`, `nancy`
    - Deploy: SSH zu Self-Hosted Server
  - Mobile CI/CD:
    - GitHub Actions Workflow: `.github/workflows/mobile.yml`
    - Lint: ESLint, Prettier
    - Test: Jest
    - Build: Expo EAS Build
    - Deploy: Expo EAS Submit
  - Environment Management:
    - Staging und Production
    - Secrets Management (GitHub Secrets)
  - Automatische Deployments:
    - Main Branch → Staging
    - Tag → Production
  - Tests

  **Must NOT do**:
  - Keine manuellen Deployments (alles automatisiert)
  - Keine Secrets in Code

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: DevOps Setup
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: T4.7
  - **Blocked By**: None

  **References**:
  - GitHub Actions: CI/CD
  - Expo EAS: Mobile Builds
  - Docker: Container Builds

  **Acceptance Criteria**:
  - [ ] Backend CI/CD Workflow
  - [ ] Mobile CI/CD Workflow
  - [ ] Lint, Test, Build, Deploy Stages
  - [ ] Security Scanning
  - [ ] Automatisches Deployment
  - [ ] Secrets Management

  **QA Scenarios**:
  ```
  Scenario: CI/CD Pipeline läuft erfolgreich
    Tool: GitHub Actions
    Steps:
      1. Push zu main Branch
      2. Pipeline startet
      3. Alle Stages erfolgreich
    Expected Result: Grüne Checks, Deployment erfolgreich
    Evidence: GitHub Actions Logs
  ```

  **Commit**: YES
  - Message: `ci: setup github actions for backend and mobile`
  - Files: `.github/workflows/*.yml`

- [x] 4.7. E2E Tests & Performance

  **What to do**:
  - Backend Tests:
    - Integration Tests mit Testcontainers (PostgreSQL, Redis)
    - API Contract Tests mit OpenAPI
    - Load Tests mit k6 (100 concurrent users)
  - Mobile Tests:
    - E2E Tests mit Maestro:
      - Onboarding Flow
      - Auth Flow
      - Todo CRUD Flow
      - Connection Flow
      - Gamification Flow
  - Performance Tests:
    - API Response Time < 200ms (95th percentile)
    - Mobile App Start < 3s
    - Sync Duration < 1s
  - Security Tests:
    - `gosec` für Go Code
    - `nancy` für Dependencies
    - OWASP ZAP Scan (optional)
  - Dokumentation:
    - API Dokumentation mit Swagger/OpenAPI
    - Test-Report
  - Tests

  **Must NOT do**:
  - Keine manuellen Tests (alles automatisiert)
  - Keine flaky Tests

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Testing, Performance
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: T4.8
  - **Blocked By**: Alle Implementation Tasks

  **References**:
  - Testcontainers: Integration Tests
  - k6: Load Testing
  - Maestro: Mobile E2E
  - gosec: Security Scanning

  **Acceptance Criteria**:
  - [ ] Integration Tests mit Testcontainers
  - [ ] E2E Tests mit Maestro
  - [ ] Load Tests: 100 concurrent users
  - [ ] Performance innerhalb Limits
  - [ ] Security Scans erfolgreich
  - [ ] API Dokumentation

  **QA Scenarios**:
  ```
  Scenario: E2E Test Suite läuft erfolgreich
    Tool: Maestro
    Steps:
      1. maestro test tests/onboarding.yaml
      2. maestro test tests/todo-flow.yaml
      3. maestro test tests/connection-flow.yaml
    Expected Result: Alle Tests PASS
    Evidence: Maestro Test Report
  ```

  **Commit**: YES
  - Message: `test: add e2e tests and performance benchmarks`
  - Files: `tests/`, Load Test Scripts

- [x] 4.8. Security Hardening

  **What to do**:
  - SSL/TLS Konfiguration:
    - Nginx mit Let's Encrypt
    - HTTPS Redirect
    - HSTS Header
    - SSL Labs A+ Rating
  - Security Headers:
    - Content Security Policy (CSP)
    - X-Content-Type-Options: nosniff
    - X-Frame-Options: DENY
    - Referrer-Policy
  - Input Validation:
    - `go-playground/validator` für alle Inputs
    - SQL Injection Prevention (parameterized queries)
    - XSS Prevention
  - Authentication Security:
    - JWT Secrets rotieren
    - API Keys in DB (nicht Logs)
    - Password Policy (min 8 Zeichen, komplexität)
  - Rate Limiting (optional aber empfohlen):
    - 100 req/min default
    - 5 req/min für Auth Endpoints
  - Security Audit:
    - `gosec` Scan
    - Manuelle Code Review
    - OWASP Top 10 Check
  - Tests

  **Must NOT do**:
  - Keine unverschlüsselten Verbindungen
  - Keine Secrets in Logs

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Security, DevOps
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: FINAL Tasks
  - **Blocked By**: T4.5, T4.7

  **References**:
  - OWASP Top 10
  - SSL Labs Best Practices
  - gosec: Security Scanner

  **Acceptance Criteria**:
  - [ ] SSL/TLS konfiguriert (A+ Rating)
  - [ ] Alle Security Headers gesetzt
  - [ ] Input Validation implementiert
  - [ ] Authentication sicher
  - [ ] Rate Limiting (optional)
  - [ ] Security Audit bestanden
  - [ ] Tests

  **QA Scenarios**:
  ```
  Scenario: SSL Labs Scan gibt A+
    Tool: SSL Labs
    Steps:
      1. Führe SSL Labs Scan durch
      2. Prüfe Rating
    Expected Result: A+ Rating
    Evidence: SSL Labs Report
  ```

  **Commit**: YES
  - Message: `security: harden application with ssl, headers, and validation`
  - Files: `nginx.conf`, Security Updates

---

## Final Verification Wave

- [x] F1. Plan Compliance Audit — `oracle`

  **What to do**:
  - Lese den gesamten Plan durch
  - Für jedes "Must Have": Verifiziere Implementation existiert
  - Für jedes "Must NOT Have": Suche Codebase nach verbotenen Patterns
  - Prüfe Evidence Files in `.sisyphus/evidence/`
  - Vergleiche Deliverables gegen Plan
  
  **Acceptance Criteria**:
  - [ ] Must Have [N/N] - Alle implementiert
  - [ ] Must NOT Have [N/N] - Keine verbotenen Patterns
  - [ ] Tasks [N/N] - Alle abgeschlossen
  - [ ] VERDICT: APPROVE

- [x] F2. Code Quality Review — `unspecified-high`

  **What to do**:
  - `tsc --noEmit` für Mobile
  - `golangci-lint` für Backend
  - `go test` und `npm test`
  - Review auf: as any, leere catches, console.log, auskommentierter Code
  - AI-Slop Check: Zu viele Kommentare, Über-Abstraktion, generische Namen
  
  **Acceptance Criteria**:
  - [ ] Build [PASS]
  - [ ] Lint [PASS]
  - [ ] Tests [N pass]
  - [ ] VERDICT

- [x] F3. Real Manual QA — `unspecified-high`

  **What to do**:
  - Clean state, führe alle QA Scenarios aus
  - Cross-Task Integration testen
  - Edge Cases: Empty state, invalid input, rapid actions
  - Speichere Evidence zu `.sisyphus/evidence/final-qa/`
  
  **Acceptance Criteria**:
  - [ ] Scenarios [N/N pass]
  - [ ] Integration [N/N]
  - [ ] Edge Cases [N tested]
  - [ ] VERDICT

- [x] F4. Scope Fidelity Check — `deep`

  **What to do**:
  - Vergleiche jede Task mit tatsächlichem Diff
  - Prüfe auf Scope Creep
  - Cross-Task Contamination checken
  
  **Acceptance Criteria**:
  - [ ] Tasks [N/N compliant]
  - [ ] Contamination [CLEAN]
  - [ ] Unaccounted [CLEAN]
  - [ ] VERDICT

---

## Commit Strategy

### Branch Structure
```
main                    # Production (protected)
├── develop             # Integration branch
│   ├── feature/auth    # JWT + API Keys
│   ├── feature/users   # User management
│   ├── feature/todos   # CRUD + optimistic locking
│   ├── feature/sync    # Conflict resolution + WatermelonDB sync
│   ├── feature/connect # 1:1 connections + invites
│   ├── feature/gamify  # Points/Levels/Badges + anti-cheat
│   ├── feature/push    # Notification queue
│   └── feature/admin   # Health checks + observability
└── hotfix/*            # Emergency production fixes
```

### Commit Convention
```
type(scope): subject

types: feat, fix, test, refactor, docs, security, perf
```

---

## Success Criteria

### Verification Commands
```bash
# Backend
go test ./... -race -coverprofile=coverage.out  # >80% coverage
go build ./...
curl http://localhost:8080/health/ready  # {"status":"ready"}

# Mobile
npm test
npx tsc --noEmit
maestro test tests/onboarding.yaml  # PASS

# Security
gosec ./...  # No high severity issues
nancy go.sum  # No vulnerable dependencies
# SSL Labs: A+ rating
```

### Final Checklist
- [ ] All "Must Have" present and working
- [ ] All "Must NOT Have" absent
- [ ] All tests pass
- [ ] Code coverage > 80%
- [ ] Security audit passed
- [ ] Performance: p95 < 200ms
- [ ] CI/CD pipeline green