# -----------------------------------------------------------------------------
# STAGE 1: Build del Frontend (Vue.js)
# -----------------------------------------------------------------------------
FROM node:20-slim AS frontend-builder

# Impostiamo la cartella di lavoro
WORKDIR /app/webui

# Copiamo i file di configurazione delle dipendenze
COPY webui/package.json webui/yarn.lock ./

# Installiamo le dipendenze
RUN yarn install --frozen-lockfile

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
# CGO_ENABLED=1 serve perché usiamo go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -o webapi ./cmd/webapi

# -----------------------------------------------------------------------------
# STAGE 3: Immagine Finale (Runtime)
# -----------------------------------------------------------------------------
# Usiamo un'immagine Debian leggera (necessaria per SQLite/CGO)
FROM debian:bookworm-slim

# Esponiamo la porta del server
EXPOSE 3000

# Impostiamo la cartella di lavoro
WORKDIR /app

# 1. Copiamo l'eseguibile dal builder del backend
COPY --from=backend-builder /app/webapi .

# 2. Copiamo la cartella 'dist' compilata dal builder del frontend
#    La mettiamo esattamente dove il backend si aspetta di trovarla (./webui/dist)
COPY --from=frontend-builder /app/webui/dist ./webui/dist

# Comando di avvio
CMD ["./webapi"]