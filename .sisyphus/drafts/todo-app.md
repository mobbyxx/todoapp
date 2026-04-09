# Draft: Kollaborative Todo-App mit Gamification

## Erstellt
2025-01-08

## Vision
Eine mobile Todo-App für mehrere Personen, bei der man sich gegenseitig Todos schreiben kann, mit Gamification-Elementen und individuell festlegbaren Belohnungen.

## Requirements (confirmed)

### Tech-Stack
- **Mobile App**: Echte native App (nicht Web/PWA)
- **Backend**: Go
- **Datenbank**: PostgreSQL
- **Hosting**: Self-hosted Server
- **API**: REST mit API Keys

### Funktionalität
- **Multi-User**: 1:1 Verbindungen zwischen Usern
- **Todo-Sharing**: Standardmäßig geteilte Todos (nicht privat)
- **Gamification**: Punkte/Sterne-System mit Levels/Badges
- **Belohnungen**: Custom, individuell wählbare Belohnungen
- **Benachrichtigungen**: Push-Benachrichtigungen
- **Automatisierung**: REST API für externe Integrationen (z.B. Chatbot-Assistent)

### Scope
- **Ziel**: Produktiv einsetzbare Anwendung
- **Umfang**: Komplettes Projekt (kein MVP)

## Entscheidungen (final)

### Mobile Framework
**React Native mit Expo** (basierend auf Recherche)
- Besser für REST APIs (TanStack Query)
- Einfacheres Push Setup
- Größerer Developer-Pool
- Kleinere App-Größe

### Authentifizierung & Verbindung
**BOTH: Einladungslink (A) UND QR-Code (B)**
- Deep Linking für Einladungslinks
- In-App QR-Code Scanner
- Verbindungsanfrage muss bestätigt werden

### Sicherheit
- Domain: Vorhanden
- SSL: User kann einrichten
- Rate Limiting: Nicht kritisch für MVP
- DB-Verschlüsselung: Empfohlen (Prometheus entscheidet)

### Skalierung
- Initial: < 100 User
- Potenzial: > 100 User möglich
- Architektur: Auf Wachstum ausgelegt
  - PostgreSQL mit Partitioning-Strategie
  - Connection Pooling
  - Caching-Layer (Redis)
  - Horizontale Skalierung vorbereitet

## Technische Entscheidungen (pending)

### Mobile App
- **Option A**: Flutter (eine Codebase, iOS + Android)
- **Option B**: React Native (JavaScript/TypeScript)
- **Option C**: Native (Swift + Kotlin, separate Codebases)

### Backend-Architektur
- Go mit Standard HTTP-Router (chi, gin, oder stdlib)
- PostgreSQL mit GORM oder sqlx
- JWT für Authentifizierung
- API Key Management für externe Integrationen

### Datenbank-Schema (initial thoughts)
- users
- connections (1:1 Beziehungen)
- todos
- rewards
- gamification_stats (punkte, level, badges)
- push_tokens

## Research Findings (pending)

Warte auf Ergebnisse von:
- Go Backend Patterns für Mobile APIs
- Mobile Framework Vergleich
- PostgreSQL Schema für Todos + Gamification
- Push Notification Best Practices

## Nächste Schritte
1. Kläre offene Fragen mit User
2. Führe Research durch
3. Konsultiere Metis für Gap-Analyse
4. Erstelle detaillierten Work Plan
5. High Accuracy Review mit Momus
