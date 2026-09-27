# BIA Energy AI — AI Energy Management Platform

MVP de gestión energética con IA: **DATOS → ANÁLISIS → ANOMALÍA → EXPLICACIÓN → PRIORIZACIÓN → ACCIÓN**.

| Proyecto | Stack | Puerto | Descripción |
|---|---|---|---|
| [`BackendBia/`](BackendBia/README.md) | Go 1.24 · chi · JWT | `8080` | API REST + motor de IA de anomalías |
| [`FrontendBia/`](FrontendBia/README.md) | React 19 · Vite · TypeScript · React Router · React Query · Recharts | `3000` | Aplicación web que consume la API |

**Demo en video:** [ver demo (5–10 min)](PEGAR_AQUI_EL_LINK_DEL_VIDEO)

**Requisito:** solo **Docker** (con Docker Compose).
**Usuario demo:** `demo@bia.energy` / `demo123`

---

## Cómo se conectan

El **backend** es solo una API (responde JSON en `/api`, no tiene pantallas). El **frontend** es la aplicación web: nginx sirve las pantallas y reenvía las llamadas `/api` al backend.

```
Navegador ──► Frontend :3000 ──┬── /        → pantallas (React)
                               └── /api/... → Backend :8080 → JSON
```

Hay **dos formas** de levantar el proyecto. Usa **una u otra**, no las dos al mismo tiempo.

---

## Opción 1 — Todo junto (un solo comando)

Desde la carpeta `BiaEnergy`:

```bash
docker compose up --build
```

Docker levanta los dos servicios en el orden correcto: primero el backend y, cuando está sano, el frontend. Los logs de ambos salen en la misma terminal, identificados como `bia-backend` y `bia-frontend`.

Para detenerlo usa `Ctrl + C`, o:

```bash
docker compose down        # detener (conserva análisis y notas)
docker compose down -v     # detener y borrar análisis y notas (demo desde cero)
```

---

## Opción 2 — Cada uno en su terminal

Útil para ver los logs de cada servicio por separado o reiniciar uno sin tocar el otro. Cada proyecto tiene su propio `docker-compose.yml`.

**Terminal 1 — Backend** (levantar primero)

```bash
cd BiaEnergy/BackendBia
docker compose up --build
```

Espera a ver `API escuchando address=:8080`.

**Terminal 2 — Frontend**

```bash
cd BiaEnergy/FrontendBia
docker compose up --build
```

El frontend encuentra al backend en el puerto 8080 de tu equipo a través de `host.docker.internal`. Para detener cada servicio, usa `Ctrl + C` en su terminal o `docker compose down` en su carpeta.

---

## Verificar que funciona

| Qué | URL | Resultado esperado |
|---|---|---|
| **Aplicación** | <http://localhost:3000> | Pantalla de inicio de sesión |
| API | <http://localhost:8080/api/health> | `{"meters":12,"status":"ok"}` |

> `http://localhost:8080/` responde `404 page not found`. Es lo esperado: el backend solo expone `/api`, y la aplicación está en el puerto **3000**.

---

## Cambiar de una opción a la otra

Las dos opciones usan los mismos nombres de contenedor y los mismos puertos. Antes de cambiar, detén la que esté corriendo:

```bash
docker rm -f bia-backend bia-frontend
```

Los análisis y notas guardados son independientes en cada opción, porque cada una usa su propio volumen de Docker.

---

## Problemas frecuentes

| Síntoma | Solución |
|---|---|
| `port is already allocated` o `container name ... already in use` | Hay otra instancia corriendo: `docker rm -f bia-backend bia-frontend` |
| `localhost:8080` muestra pantallas antiguas | Es un contenedor de la primera versión: `docker rm -f bia-energy-ai` |
| `open /app/data/readings.csv: permission denied` | Reconstruye sin caché: `docker compose build --no-cache && docker compose up` |
| Al iniciar sesión: *"No se pudo conectar con la API"* | El backend no está corriendo: revisa <http://localhost:8080/api/health> (en la opción 2, levanta primero el backend) |
| Los cambios de código no se ven | Vuelve a construir: `docker compose up --build` |

---

## Recorrido de la demo

**Login → Dashboard → M-109 → Run AI Analysis → Ver investigación → Iniciar investigación → Anomalías IA**

1. **Dashboard:** KPIs y consumo total frente al baseline. Todavía no hay análisis.
2. **M-109:** el consumo se duplica el 12/09 a las 14:00 y el factor de potencia cae de 0,94 a 0,74.
3. **Run AI Analysis:** 7 etapas en vivo. Resultado: *4 anomalías · 2 prioritarias*.
4. **Investigación:** explicación, comparación contra baseline, variables, evidencia, confianza desglosada y JSON del modelo.
5. **Acción:** iniciar la investigación y agregar una nota (queda guardada).
6. **Anomalías IA:** M-109 real/alta, M-112 calidad de datos/alta, M-104 explicable/media y M-106 falso positivo/baja.

---

## Desarrollo sin Docker (opcional)

Con **Go ≥ 1.24** y **Node ≥ 20**, en dos terminales:

```bash
cd BiaEnergy/BackendBia && go run ./cmd/api
cd BiaEnergy/FrontendBia && npm install && npm run dev
```

La aplicación queda en <http://localhost:5173>, con recarga automática al guardar.

Detalle técnico de cada proyecto: [`BackendBia/README.md`](BackendBia/README.md) · [`FrontendBia/README.md`](FrontendBia/README.md)
