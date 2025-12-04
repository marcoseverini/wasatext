#!/bin/bash

# Costruisci le immagini Docker
if [ "$1" == "front" ]; then
    echo "Building frontend..."
    docker build -t wasa-text-frontend:latest -f Dockerfile.frontend .
elif [ "$1" == "back" ]; then
    echo "Building backend..."
    docker build -t wasa-text-backend:latest -f Dockerfile.backend .
else
    echo "Building frontend and backend..."
    docker build -t wasa-text-frontend:latest -f Dockerfile.frontend .
    docker build -t wasa-text-backend:latest -f Dockerfile.backend .
fi