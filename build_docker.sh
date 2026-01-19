#!/bin/bash

# Costruisci le immagini Docker partendo dai Dockerfile 
if [ "$1" == "front" ]; then
    echo "Building frontend..."
    sudo docker build -t wasa-text-frontend:latest -f Dockerfile.frontend .
elif [ "$1" == "back" ]; then
    echo "Building backend..."
    sudo docker build -t wasa-text-backend:latest -f Dockerfile.backend .
else
    echo "Building frontend and backend..."
    sudo docker build -t wasa-text-frontend:latest -f Dockerfile.frontend .
    sudo docker build -t wasa-text-backend:latest -f Dockerfile.backend .
fi