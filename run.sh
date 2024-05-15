#!/bin/bash

set -e -u -o pipefail

# Initialize variables
VOLUME_NAME="autodba_postgres_data"
CONTAINER_NAME="pgautodba"
IMAGE_NAME="pgautodba-image"
HOST_PORT=8081
CONTAINER_PORT=8080
RECREATE_VOLUME=false
ENV_TYPE="dev"  # Default environment
DOCKERFILE="Dockerfile"
ENV_FILE=".env.dev"

# Function to display usage information
usage() {
    echo "Usage: $0 [--recreate] [--env dev|prod]"
    echo "  --recreate      Recreate Docker volume even if it exists"
    echo "  --env           Specify the environment ('dev' for development or 'prod' for production)"
    exit 1
}

# Parse command-line options
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --recreate)
            RECREATE_VOLUME=true
            shift
            ;;
        --env)
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                ENV_TYPE="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                exit 1
            fi
            ;;
        *)
            echo "Invalid argument: $1" >&2
            usage
            ;;
    esac
done

# Define Dockerfile and .env file based on the environment type
if [ "$ENV_TYPE" = "prod" ]; then
    DOCKERFILE="Dockerfile.prod"
    ENV_FILE=".env.prod"
elif [ "$ENV_TYPE" = "github" ]; then
    DOCKERFILE="Dockerfile"
    ENV_FILE=".env.github"
fi

# Create or recreate Docker volume
if [ "$RECREATE_VOLUME" = true ]; then
    echo "Recreating Docker volume '$VOLUME_NAME'..."
    docker volume rm "$VOLUME_NAME" || echo "No existing volume to remove."
    docker volume create "$VOLUME_NAME"
else
    if ! docker volume inspect "$VOLUME_NAME" > /dev/null 2>&1; then
        echo "Creating Docker volume '$VOLUME_NAME'..."
        docker volume create "$VOLUME_NAME"
    else
        echo "Docker volume '$VOLUME_NAME' already exists."
    fi
fi

# Build Docker image
echo "Building Docker image using Dockerfile: $DOCKERFILE ..."
docker build -t "$IMAGE_NAME" -f "$DOCKERFILE" .

echo "Stopping and removing existing container '$CONTAINER_NAME'..."
docker stop "$CONTAINER_NAME" > /dev/null 2>/dev/null || true
docker rm "$CONTAINER_NAME" > /dev/null 2>/dev/null || true

# Run the container
echo "Running Docker container '$CONTAINER_NAME' on port $HOST_PORT..."
docker run --name "$CONTAINER_NAME" \
    -v "$VOLUME_NAME":/var/lib/postgresql/data \
    -p "$HOST_PORT":"$CONTAINER_PORT" \
    --env-file "$ENV_FILE" \
    "$IMAGE_NAME"
