# -----------------------------------------------------------------------------
# STAGE 1: Build del Frontend (Vue.js)
# -----------------------------------------------------------------------------
FROM node:20-slim AS frontend-builder

# Abilitiamo Corepack per supportare Yarn 4.5.0
RUN corepack enable

# Impostiamo la cartella di lavoro
WORKDIR /app/webui

# Copiamo i file di configurazione delle dipendenze
COPY webui/package.json webui/yarn.lock ./

# Installiamo le dipendenze
# --- MODIFICA QUI: Rimosso '--frozen-lockfile' per permettere aggiornamenti al lockfile ---
RUN yarn install
# -----------------------------------------------------------------------------------------

# Copiamo tutto il codice sorgente del frontend
COPY webui .

# Eseguiamo la build (creerà la cartella 'dist')
RUN yarn build

# -----------------------------------------------------------------------------
# STAGE 2: Build del Backend (Go)
# -----------------------------------------------------------------------------
FROM golang:1.23 AS backend-builder

# Impostiamo la cartella di lavoro
WORKDIR /app

# Copiamo i file del progetto Go
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY service/ service/
COPY cmd/ cmd/

# Compiliamo il backend
RUN CGO_ENABLED=1 GOOS=linux go build -o webapi ./cmd/webapi

# -----------------------------------------------------------------------------
# STAGE 3: Immagine Finale (Runtime)
# -----------------------------------------------------------------------------
FROM debian:bookworm-slim

# Esponiamo la porta del server
EXPOSE 3000

# Impostiamo la cartella di lavoro
WORKDIR /app

# 1. Copiamo l'eseguibile dal builder del backend
COPY --from=backend-builder /app/webapi .

# 2. Copiamo la cartella 'dist' compilata dal builder del frontend
COPY --from=frontend-builder /app/webui/dist ./webui/dist

# Comando di avvio
CMD ["./webapi"]