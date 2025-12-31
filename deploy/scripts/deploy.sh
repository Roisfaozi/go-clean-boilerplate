#!/bin/bash

# Configuration
DOCKER_COMPOSE_FILE="docker-compose.prod.yml"
NGINX_CONTAINER="nginx_gateway"
UPSTREAM_CONF="./deploy/nginx/upstream.conf"

# Determine current active environment
if grep -q "app-blue" "$UPSTREAM_CONF"; then
    CURRENT="blue"
    NEW="green"
else
    CURRENT="green"
    NEW="blue"
fi

echo "🔵 Current active environment: $CURRENT"
echo "🟢 Deploying to: $NEW"

# 1. Database Migration
# We run migration before starting the new app to ensure schema is ready
echo "💾 Running database migrations..."
# Assuming we use a migration tool or the app has a migrate flag
# Here we use the migrate/migrate docker image for a clean approach
docker compose -f $DOCKER_COMPOSE_FILE run --rm app-$NEW ./main -migrate=up || {
    echo "❌ Migration failed! Aborting deployment."
    exit 1
}

# 2. Bring up the new container
echo "🚀 Starting app-$NEW..."
docker compose -f $DOCKER_COMPOSE_FILE up -d --build app-$NEW

# 3. Wait for healthcheck
echo "⏳ Waiting for app-$NEW to be healthy..."
RETRIES=0
MAX_RETRIES=30
while [ $RETRIES -lt $MAX_RETRIES ]; do
    HEALTH_STATUS=$(docker inspect --format='{{.State.Health.Status}}' app-$NEW 2>/dev/null)
    if [ "$HEALTH_STATUS" == "healthy" ]; then
        echo "✅ app-$NEW is healthy!"
        break
    fi
    echo "   ...status: $HEALTH_STATUS. Retrying in 2s..."
    sleep 2
    RETRIES=$((RETRIES+1))
done

if [ $RETRIES -eq $MAX_RETRIES ]; then
    echo "❌ Deployment failed: app-$NEW failed healthcheck."
    echo "⚠️  Rolling back... Stopping app-$NEW."
    docker compose -f $DOCKER_COMPOSE_FILE stop app-$NEW
    exit 1
fi

# 4. Switch Traffic
echo "🔄 Switching Nginx traffic to app-$NEW..."
echo "upstream backend { server app-$NEW:8080; }" > "$UPSTREAM_CONF"

# 5. Reload Nginx
echo "Hz Reloading Nginx..."
docker exec $NGINX_CONTAINER nginx -s reload

# 6. Stop old container
echo "🛑 Stopping old environment: app-$CURRENT..."
docker compose -f $DOCKER_COMPOSE_FILE stop app-$CURRENT

echo "🎉 Deployment Complete! Active: $NEW"