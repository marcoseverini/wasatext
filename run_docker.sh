#!/bin/bash

# Avvia i container Docker partendo dalle immagini create tramite build_docker.sh
sudo docker run --rm -p 3000:3000 wasa-text-backend:latest & # Avvio del backend nella porta 3000
sudo docker run --rm -p 8080:80 wasa-text-frontend:latest & # Avvio del frontend nella porta 8080

wait