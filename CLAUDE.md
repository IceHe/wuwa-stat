# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

This repository is a web-based stats tracking tool for the game Wuthering Waves (鸣潮).

Current scope:
- Tacet Zone (无音区) drops
- Resonator ascension material drops
- Resonance domain (凝素领域) drops

Use "鸣潮统计网页工具" as the short product name in user-facing docs.

## Technology Stack

- **Backend**: Go + net/http + database/sql + PostgreSQL
- **Frontend**: Vue 3 + TypeScript + Element Plus + Vite
- **Auth**: external token validation service

## Common Commands

### Backend

```bash
cd backend

# Initialize database (creates tables)
go run ./cmd/initdb

# Start development server
go run ./cmd/server
```

### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Optional mirror for China
npm install --registry=https://registry.npmmirror.com

# Start development server
npm run dev

# Build for production
npm run build
```

## Architecture

### Backend Structure

```text
backend/
├── cmd/
│   ├── server/         # HTTP service entrypoint
│   └── initdb/         # Database initialization entrypoint
├── internal/
│   ├── api/            # HTTP handlers, auth, request/response types
│   ├── config/         # .env loading
│   └── db/             # DB connection and schema management
└── .env                # Runtime configuration
```

### Frontend Structure

```text
frontend/
├── src/
│   ├── main.ts         # Vue app entry point
│   ├── App.vue         # Main app with auth + tab navigation
│   ├── api/index.ts    # Axios API client and auth token handling
│   └── components/
│       ├── Tacet*.vue       # Tacet records UI
│       ├── Ascension*.vue   # Ascension records UI
│       └── Resonance*.vue   # Resonance records UI
└── vite.config.ts      # Vite config with API proxy
```

## Database Schema

Tables:
- `tacet_records`
- `ascension_records`
- `resonance_records`

## API Endpoints

- Tacet:
  - `POST /api/tacet_records`
  - `GET /api/tacet_records`
  - `GET /api/stats`
  - `GET /api/detailed-stats`
  - `GET /api/player-ids`
  - `DELETE /api/tacet_records/{id}`
- Ascension:
  - `POST /api/ascension-records`
  - `GET /api/ascension-records`
  - `GET /api/ascension-detailed-stats`
  - `GET /api/ascension-player-ids`
  - `DELETE /api/ascension-records/{id}`
- Resonance:
  - `POST /api/resonance-records`
  - `GET /api/resonance-records`
  - `GET /api/resonance-detailed-stats`
  - `GET /api/resonance-player-ids`
  - `DELETE /api/resonance-records/{id}`
- Auth:
  - `GET /api/auth/me`

## Key Configuration

- Backend listens on port 8000
- Frontend dev server listens on port 5173
- Database connection configured in `backend/.env` via `DATABASE_URL`
- CORS frontend URL configured via `FRONTEND_URL`
- Auth is delegated to an external service, default `http://127.0.0.1:8080`
- Most API routes require a token with `view`, `edit`, or `manage` permissions

## Production Deployment (systemd)

Services are configured as systemd units for permanent background running:

```bash
systemctl start wuwa-stat-backend
systemctl start wuwa-stat-frontend
systemctl stop wuwa-stat-backend
systemctl stop wuwa-stat-frontend

journalctl -u wuwa-stat-backend -f
journalctl -u wuwa-stat-frontend -f

systemctl restart wuwa-stat-backend
systemctl restart wuwa-stat-frontend
```

Service files:
- `/etc/systemd/system/wuwa-stat-backend.service`
- `/etc/systemd/system/wuwa-stat-frontend.service`
