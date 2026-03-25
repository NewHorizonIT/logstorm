# LogStorm Dashboard

A modern, high-performance log analytics dashboard for **LogStorm** — designed for real-time monitoring, log exploration, and system observability.

---

## Overview

LogStorm Dashboard is a web-based UI that allows developers and operators to:

- Search and explore logs in real-time
- Visualize system metrics and trends
- Monitor errors and anomalies
- Manage ingestion and API keys

Inspired by tools like Grafana and Kibana, but tailored for **log-first observability systems**.

---

## Architecture

This project follows a **modular + scalable frontend architecture**:

```
src/
│
├── app/                # App bootstrap (providers, router)
│   ├── App.tsx
│   ├── routes.tsx
│
├── modules/            # Feature modules
│   ├── dashboard/
│   ├── logs/
│   ├── metrics/
│   ├── alerts/
│   ├── api-keys/
│
├── components/         # Reusable UI components
│   ├── ui/
│   ├── common/
│
├── layouts/            # Layout system
│   ├── MainLayout.tsx
│   ├── Sidebar.tsx
│   ├── Topbar.tsx
│
├── services/           # API layer (axios instances)
├── store/              # Global state (Zustand)
├── hooks/              # Custom hooks
├── utils/              # Helpers
├── types/              # TypeScript types
```

---

## Tech Stack

- React (with Vite)
- TailwindCSS (UI styling)
- Zustand (state management)
- Recharts
- Axios

---

## Core Features

### Dashboard

- System overview (logs/sec, error rate, services)
- Real-time charts
- Top services & recent errors

---

### Logs Explorer (Core Feature)

- Full-text search
- Filter by:
  - service
  - log level
  - time range

- Expandable log rows
- JSON log viewer

---

### Log Detail

- Full structured log
- Trace ID navigation
- Related logs
- Timeline view

---

### Metrics & Analytics

- Query builder (log aggregation)
- Charts (line, bar, pie)
- Saved queries

---

### Alerts

- Rule-based alert system
- Alert history
- Notification channels (future)

---

### API Key Management

- Create & revoke API keys
- Permission control
- Multi-tenant support

---

## Performance Optimizations

- Virtualized log table (for large datasets)
- Debounced search queries
- Lazy loading modules
- Memoized components

---

## Data Flow

```
User Action → UI Component → Hook → Service (API) → Backend
                                      ↓
                                 Global Store
                                      ↓
                                   UI Update
```

---

## API Integration

The frontend communicates with LogStorm backend services:

- Log ingestion API
- Query API (search + aggregation)
- Metrics API
- Auth & API key management

---

## Getting Started

### 1. Install dependencies

```bash
npm install
```

### 2. Run development server

```bash
npm run dev
```

### 3. Build for production

```bash
npm run build
```

---

## Environment Variables

```env
VITE_API_BASE_URL=http://localhost:8080
```

---

## Future Improvements

- Distributed tracing (Jaeger-like view)
- AI log summarization
- Anomaly detection
- WebSocket real-time streaming logs

---

## Design Philosophy

- **Log-first observability** (logs are the source of truth)
- **Developer-friendly UX**
- **High performance at scale**
- **Modular & extensible architecture**
