#!/bin/bash

# Start node1
docker-compose up -d node1 --build
echo "Started node1, waiting 2 seconds..."
sleep 2

# Start node2
docker-compose up -d node2 --build
echo "Started node2, waiting 2 seconds..."
sleep 2

# Start node3
docker-compose up -d node3 --build
echo "Started node3"

# Optional: Display logs
docker-compose logs -f
