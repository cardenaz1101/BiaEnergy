# BIA Energy AI — Frontend

SPA en **React 19 + Vite + TypeScript** que consume la API de `../BackendBia`.

## Stack

| Necesidad | Librería |
|---|---|
| Rutas, rutas protegidas, filtros en la URL | [`react-router`](https://reactrouter.com) 7 |
| Datos del servidor: caché, estados de carga, polling del análisis, invalidación | [`@tanstack/react-query`](https://tanstack.com/query) 5 |
| Cliente HTTP con interceptores (token y 401) | [`axios`](https://axios-http.com) |
| Gráficas (banda de rango normal, tramos anómalos, eventos, tooltip) | [`recharts`](https://recharts.org) 3 |
| Íconos (set Lucide) | [`react-icons`](https://react-icons.github.io/react-icons) |
| Clases condicionales | [`clsx`](https://github.com/lukeed/clsx) |

## Cómo ejecutarlo

Desde `BiaEnergy`: `docker compose up --build` → <http://localhost:3000> (**demo@bia.energy / demo123**).

En desarrollo, con Node ≥ 20 y el backend en `:8080`:

```bash
npm install
npm run dev        # http://localhost:5173 (Vite reenvía /api al backend)
npm run typecheck
npm run build
```

| Variable | Dónde | Descripción |
|---|---|---|
| `VITE_API_URL` | build | URL base de la API. Por defecto `/api` (mismo origen) |
| `API_PROXY_TARGET` | `npm run dev` | Backend al que Vite reenvía `/api` |
| `API_UPSTREAM` | contenedor nginx | Backend al que nginx reenvía `/api` |

## Estructura

```
src/
  api/
    http.ts           instancia de axios, interceptores y mensajes de error
    endpoints.ts      una función por endpoint de la API
    queries.ts        hooks de React Query (useMeters, useAnomaly, useStartAnalysis, ...)
    types.ts          contrato de la API
  auth/               sesión, AuthProvider y RequireAuth
  hooks/
    AnalysisProvider.tsx  estado global del análisis: polling por etapa y recarga al terminar
  components/
    layout/           AppLayout (menú), Page (cabecera), AnalysisPipeline, RunAnalysisButton
    charts/           ConsumptionChart, MetricChart, Sparkline, colores del tema
    anomalies/        AnomalyQueueItem, EventItem, PriorityBadge
    ui/               Chips, KpiCard, estados de carga, error y vacío
  lib/                formato es-CO, etiquetas y datos de gráficas
  pages/              Login, Dashboard, Meters, MeterDetail, Anomalies, Investigation
    investigation/    ActionPanel, ConfidencePanel, VariablesTable, EvidenceList, BaselineComparison
```

## Rutas

| Ruta | Pantalla |
|---|---|
| `/login` | Inicio de sesión (vuelve a la página solicitada) |
| `/` | Dashboard: KPIs, consumo total vs baseline, cola de prioridad y estado de medidores |
| `/meters?status=&q=&sort=&order=` | Medidores con filtros, búsqueda y orden guardados en la URL |
| `/meters/:meterId` | Detalle: consumo vs esperado (horario/diario), voltaje, corriente, FP, eventos y hallazgo |
| `/anomalies` | Anomalías priorizadas con filtro por tipo |
| `/anomalies/:anomalyId` | Investigación: explicación, baseline, variables, evidencia, confianza, acción y notas |

## Decisiones

- **Mismo origen:** el navegador siempre llama a `/api`. En desarrollo lo reenvía Vite y en Docker lo hace nginx.
- **React Query como fuente de verdad de los datos del servidor:** al terminar un análisis se invalidan las consultas y cada pantalla se actualiza sola.
- **Estados con ícono y texto, no solo color:** además hay modo oscuro según el sistema y el diseño se adapta a móvil.
