# BIA Energy AI — Backend (API)

API REST en **Go** de la plataforma de gestión energética con IA. Convierte lecturas de medidores en decisiones operativas:

**DATOS → ANÁLISIS → ANOMALÍA → EXPLICACIÓN → PRIORIZACIÓN → ACCIÓN**

El frontend es un proyecto aparte (`../FrontendBia`). Para levantar los dos juntos usa el `docker-compose.yml` de la carpeta `BiaEnergy`.

## Stack

| Necesidad | Solución |
|---|---|
| Router HTTP y middlewares (request id, logs, recover, timeout) | [`go-chi/chi`](https://github.com/go-chi/chi) |
| CORS | [`go-chi/cors`](https://github.com/go-chi/cors) |
| Autenticación | JWT HS256 con [`golang-jwt/jwt`](https://github.com/golang-jwt/jwt) |
| Logs estructurados | `log/slog` (librería estándar) |
| Motor de IA, CSV y persistencia | Librería estándar |

## Cómo ejecutarlo

```bash
docker compose up --build      # solo la API, sin instalar Go
go run ./cmd/api               # con Go ≥ 1.24
go run ./cmd/analyze -v        # motor de IA en la terminal
go run ./cmd/analyze -json     # salida JSON por medidor
```

La API queda en <http://localhost:8080/api> (`/api/health`). Usuario demo: **demo@bia.energy / demo123**.

| Variable | Por defecto | Descripción |
|---|---|---|
| `ADDR` | `:8080` | Dirección de escucha |
| `DATA_DIR` | `data` | Carpeta con `readings.csv`, `events.csv` y `meters.csv` |
| `STATE_PATH` | `data/state.json` | Estado persistido (análisis, estados y notas) |
| `CORS_ORIGINS` | `*` | Orígenes permitidos, separados por coma |
| `JWT_SECRET` | valor demo | Clave de firma de los tokens |
| `TOKEN_TTL` | `12h` | Vigencia del token |
| `DEMO_EMAIL` / `DEMO_PASSWORD` / `DEMO_NAME` | demo | Usuario de demostración |
| `ANALYSIS_STEP_DELAY` | `550ms` | Pausa visual entre etapas del pipeline |

## Arquitectura

Es una arquitectura en capas con **puertos y adaptadores**. Los servicios dependen de interfaces de repositorio, y el almacenamiento es un adaptador que se puede reemplazar sin tocar el resto. El motor de IA no hace I/O.

```
cmd/
  api/            punto de entrada de la API: configuración y armado de dependencias
  analyze/        CLI que ejecuta el motor de IA
internal/
  config/         configuración por variables de entorno
  domain/         entidades y reglas propias: Meter, Reading, Event, Anomaly, AnalysisRun, errores
  numeric/        estadística y redondeo (media, mediana, MAD, % de cambio)
  engine/         motor de IA, sin I/O
    baseline.go       perfil horario esperado por medidor
    deviation.go      tramos persistentes y transitorios de consumo
    dataquality.go    lecturas inconsistentes intermitentes
    variables.go      comparación de consumo, corriente, voltaje, FP y coherencia
    classify.go       clasificación con evidencia (real, explicable, falso positivo, calidad)
    confidence.go     pesos de confianza como constantes con nombre
    describe.go       título, razón, explicación y acción recomendada
    priority.go       score y ranking de prioridad
    engine.go         pipeline de 7 etapas
  service/        casos de uso y puertos (interfaces de repositorio)
  storage/        adaptador: carga de CSV + repositorio en memoria con snapshot JSON
  auth/           login y validación de tokens JWT
  httpapi/        rutas chi, handlers por recurso, middleware de autenticación y respuestas
data/             readings.csv, events.csv, meters.csv (nombre y sede de cada medidor)
```

Las dependencias van hacia adentro: `httpapi → service → engine/domain`, y `storage` implementa las interfaces de `service/ports.go`.

## Motor de IA

1. **Lecturas:** valida y ordena las 4.032 lecturas horarias.
2. **Baseline:** con los primeros 7 días calcula un perfil horario robusto (mediana y MAD) de consumo y corriente, más los valores normales de voltaje, factor de potencia y la coherencia `kWh / (V·I·FP)`.
3. **Detección:** busca tramos con más de 25% de desviación por hora. Si duran 24 h o más y siguen activos, los marca como persistentes. Detecta además lecturas eléctricas imposibles que aparecen intercaladas con lecturas normales (calidad de datos).
4. **Correlación:** separa un aumento de carga "sano", donde corriente y consumo suben juntos, de un cambio eléctrico (caída de FP o de voltaje).
5. **Eventos:** cruza eventos operativos a ±12 h del inicio y compara la duración declarada con la observada.
6. **Explicación:** clasifica con evidencia. La confianza se calcula como suma de factores con nombre, con tope de 0,97.
7. **Recomendación:** genera la acción y los pasos, y asigna un score de prioridad por tipo, severidad, confianza y energía en juego.

| # | Medidor | Tipo | Severidad | Confianza |
|---|---|---|---|---|
| 1 | M-109 | REAL_ANOMALY | HIGH | 0,97 |
| 2 | M-112 | DATA_QUALITY | HIGH | 0,92 |
| 3 | M-104 | EXPLAINABLE_ANOMALY | MEDIUM | 0,84 |
| 4 | M-106 | FALSE_POSITIVE | LOW | 0,97 |

Los otros 8 medidores quedan en OK. El motor no usa `expected_results.csv`.

## API

Todas las rutas van bajo `/api`. Requieren `Authorization: Bearer <token>`, salvo `POST /auth/login` y `GET /health`.

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/auth/login` | `{email, password}` → `{token, user}` |
| GET | `/auth/me` | Usuario del token |
| GET | `/dashboard/summary` | KPIs, consumo diario total, cola de prioridad y último análisis |
| GET | `/meters?status=&q=&sort=&order=` | `status`: all/normal/alert/critical · `sort`: consumption/variation/severity |
| GET | `/meters/{meterID}` | Resumen, baseline, eventos y anomalías del medidor |
| GET | `/meters/{meterID}/readings?granularity=hour\|day&from=&to=` | Lecturas con valor esperado y rango normal |
| GET | `/events?meter_id=` | Eventos operativos |
| GET | `/anomalies?type=&severity=&status=&meter_id=` | Anomalías ordenadas por prioridad |
| GET | `/anomalies/{anomalyID}` | Detalle con evidencia, variables y confianza |
| PATCH | `/anomalies/{anomalyID}` | `{status: OPEN\|INVESTIGATING\|RESOLVED\|DISMISSED, note}` |
| POST | `/ai/analyze` | Lanza el análisis: 202, o 409 si ya hay uno en curso |
| GET | `/ai/analysis`, `/ai/analysis/latest`, `/ai/analysis/{runID}` | Historial, último análisis y detalle por etapa |

```bash
T=$(curl -s -XPOST localhost:8080/api/auth/login -d '{"email":"demo@bia.energy","password":"demo123"}' | jq -r .token)
curl -s -XPOST -H "Authorization: Bearer $T" localhost:8080/api/ai/analyze
curl -s -H "Authorization: Bearer $T" localhost:8080/api/anomalies | jq '.[0] | {meter_id, type, severity, confidence, reason}'
```

## Siguientes pasos

- Tests del motor, con los 4 casos del dataset como fixtures, y de la API con `httptest`.
- Adaptador de PostgreSQL o TimescaleDB que implemente los mismos puertos de `service`.
- LLM opcional que redacte la explicación a partir de la evidencia estructurada.
