# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a stats tracking tool for the game Wuthering Waves (鸣潮), specifically tracking "Tacet Zone" (无音区) drop data - gold and purple "tubes" (密音筒) that players receive.

## Technology Stack

- **Backend**: FastAPI + SQLAlchemy + PostgreSQL
- **Frontend**: Vue 3 + TypeScript + Element Plus + Vite

## Common Commands

### Backend
```bash
cd backend

# Install dependencies
pip install -r requirements.txt

# Initialize database (creates tables)
python init_db.py

# Import sample data
python import_sample_data.py

# Start development server
uvicorn app.main:app --reload --port 8000
```

### Frontend
```bash
cd frontend

# Install dependencies (may need mirror for China)
npm install --registry=https://registry.npmmirror.com

# Start development server
npm run dev

# Build for production
npm run build
```

## Architecture

### Backend Structure
```
backend/
├── app/
│   ├── main.py         # FastAPI app entry point, CORS config
│   ├── models.py       # SQLAlchemy Record model (table: tacet_stats)
│   ├── schemas.py      # Pydantic request/response schemas
│   ├── database.py     # Database connection, SessionLocal
│   └── api/
│       └── routes.py   # API endpoints
├── init_db.py          # Database initialization script
└── .env                # Database URL configuration
```

### Frontend Structure
```
frontend/
├── src/
│   ├── main.ts         # Vue app entry point
│   ├── App.vue         # Main app with tab navigation
│   ├── api/index.ts    # Axios API client
│   └── components/
│       ├── RecordInput.vue   # Form for adding records
│       ├── RecordList.vue   # Table view with filters
│       └── StatsView.vue    # Statistics dashboard
└── vite.config.ts      # Vite config with API proxy
```

### Database Schema
Table: `tacet_stats`
- `id` (PK)
- `date` (indexed)
- `player_id` (indexed)
- `gold_tubes` - gold tube count
- `purple_tubes` - purple tube count
- `sola_level` - Sola level (1-8, default 8)
- `created_at` - timestamp

### API Endpoints
- `POST /api/records` - Batch create records
- `GET /api/records` - Query records with filters (player_id, date range, sola_level) and pagination
- `GET /api/stats` - Basic statistics (totals, averages)
- `GET /api/detailed-stats` - Stats grouped by sola level and drop combinations
- `GET /api/player-ids` - List unique player IDs
- `DELETE /api/records/{id}` - Delete a record

### API Proxy
Frontend uses Vite proxy to forward `/api` requests to `http://localhost:8000`.

## Key Configuration

- Backend listens on port 8000
- Frontend dev server on port 5173
- Database connection configured in `backend/.env` via `DATABASE_URL`
- CORS configured to allow frontend URL (default `http://localhost:5173`)

## Production Deployment (systemd)

Services are configured as systemd units for permanent background running:

```bash
# Start/Stop services
systemctl start wuwa-stat-backend
systemctl start wuwa-stat-frontend
systemctl stop wuwa-stat-backend
systemctl stop wuwa-stat-frontend

# View logs
journalctl -u wuwa-stat-backend -f
journalctl -u wuwa-stat-frontend -f

# Restart services
systemctl restart wuwa-stat-backend
systemctl restart wuwa-stat-frontend
```

Service files:
- `/etc/systemd/system/wuwa-stat-backend.service`
- `/etc/systemd/system/wuwa-stat-frontend.service`

Services are enabled to start on boot automatically.
