#!/bin/bash
set -e -u -o pipefail

# Initialize variables
# VOLUME_NAME="autodba_postgres_data"
CONTAINER_NAME="pgautodba-$USER"
IMAGE_NAME="pgautodba-image"
HOST_PORT=$(($UID + 8000))
POSTGRES_HOST_PORT=$((UID + 9000))
DOCKERFILE="Dockerfile"

# Function to display usage information
usage() {
    echo "Usage: $0"
    # echo "  --recreate      Recreate Docker volume even if it exists"
    # echo "  --env           Specify the environment ('dev' for development or 'prod' for production)"
    exit 1
}

# Parse command-line options
while [[ "$#" -gt 0 ]]; do
    case $1 in
        # --recreate)
        #     RECREATE_VOLUME=true
        #     shift
        #     ;;
        # --env)
        #     if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
        #         ENV_TYPE="$2"
        #         shift 2
        #     else
        #         echo "Error: Argument for $1 is missing" >&2
        #         exit 1
        #     fi
        #     ;;
        *)
            echo "Invalid argument: $1" >&2
            usage
            ;;
    esac
done

# TODO -- logic to handle customer db w/ retries, etc.  Then pass these through, probably as
#         a postgres URL

# TODO: These are internal implementation details so far, don't let them be customized!
# Load environment variables from .env file
# unamestr=$(uname)
# if [ "$unamestr" = 'Linux' ]; then

#   export $(grep -v '^#' $ENV_FILE | xargs -d '\n')

# elif [ "$unamestr" = 'FreeBSD' ] || [ "$unamestr" = 'Darwin' ]; then

#   export $(grep -v '^#' $ENV_FILE | xargs -0)

# fi

# TODO -- for v1, the agent DB is ephemeral, so don't do this yet.
# # Create or recreate Docker volume
# if [ "$RECREATE_VOLUME" = true ]; then
#     echo "Recreating Docker volume '$VOLUME_NAME'..."
#     docker volume rm "$VOLUME_NAME" || echo "No existing volume to remove."
#     docker volume create "$VOLUME_NAME"
# else
#     if ! docker volume inspect "$VOLUME_NAME" > /dev/null 2>&1; then
#         echo "Creating Docker volume '$VOLUME_NAME'..."
#         docker volume create "$VOLUME_NAME" || true # Ignore error, as it fails on GitHub Actions
#     else
#         echo "Docker volume '$VOLUME_NAME' already exists."
#     fi
# fi

echo "Running tests + lint ..."
docker build -f "$DOCKERFILE" . --target lint
docker build -f "$DOCKERFILE" . --target test

# Build Docker image
echo "Building Docker image using Dockerfile: $DOCKERFILE ..."
docker build -t "$IMAGE_NAME" -f "$DOCKERFILE" .

echo "Stopping and removing existing container '$CONTAINER_NAME'..."
docker stop "$CONTAINER_NAME" > /dev/null 2>/dev/null || true
docker rm "$CONTAINER_NAME" > /dev/null 2>/dev/null || true

# Run the container
echo "=============================================================="
echo ""
echo "Running Docker container: $CONTAINER_NAME"
echo ""
echo "   flask port: $HOST_PORT"
echo "      pg port: $POSTGRES_HOST_PORT"
echo ""
echo "=============================================================="

docker run --name "$CONTAINER_NAME" \
    -p "$HOST_PORT":8080 \
    -p "$POSTGRES_HOST_PORT":5432 \
    "$IMAGE_NAME"

    # -v "$VOLUME_NAME":/var/lib/postgresql/data \
    # --env-file "$ENV_FILE" \
