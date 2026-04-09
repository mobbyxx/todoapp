# Todo-App Decisions Log

## 2025-04-08: Work Session Started
- **Session ID**: ses_291c2e817ffe7qcHa2ugn5RWRs
- **Plan**: todo-app.md
- **Starting Wave**: 1 (Foundation)
- **Strategy**: Execute Wave 1 tasks in parallel where possible

## Task Breakdown Strategy
Based on dependency analysis:
1. **Parallel Batch 1**: T1.1 (Docker) + T1.6 (Mobile Setup)
2. **Parallel Batch 2**: T1.2 (Schema) + T1.3 (Go Structure)
3. **Parallel Batch 3**: T1.4 (JWT) + T1.5 (API Keys)
4. **Sequential**: T1.7 (needs T1.3, T1.4, T1.5)
5. **Sequential**: T1.8 (needs T1.6)

## Technology Stack Confirmed
- **Backend**: Go 1.21+, Chi, pgx, goose, jwt-go, bcrypt, argon2
- **Frontend**: Expo SDK 52, TanStack Query, Zustand, WatermelonDB
- **Infrastructure**: PostgreSQL 15, Redis 7, Nginx

## Security Decisions
- JWT: HS256 symmetric (simpler for self-hosted)
- API Keys: Prefix + Version + Random, Argon2id hashed
- Rate limiting: In-memory Redis counters
- CORS: Explicit origin whitelist
