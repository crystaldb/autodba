#!/bin/bash

# This script is used to backup the (Prometheus and PostgreSQL) databases inside the Docker container.

# Default values for optional parameters
BACKUP_DIR='./autodba_backups_dir'
INSTANCE_ID=0

# Function to print usage
usage() {
    echo "Usage: $0 [--backup-dir <directory>] [--instance-id <id>]"
    echo "  --backup-dir <directory>  Specify the directory to save the backup (default: ./autodba_backups_dir)"
    echo "  --instance-id <id>        Specify the instance ID (default: 0)"
    exit 1
}

# Parse command line arguments for optional parameters
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --backup-dir) BACKUP_DIR="$2"; shift ;;
        --instance-id) INSTANCE_ID="$2"; shift ;;
        -h|--help) usage ;;  # Print usage if -h or --help is provided
        *) echo "Unknown parameter passed: $1"; usage; exit 1 ;;  # Handle unknown parameters
    esac
    shift
done

# Get the Docker container ID based on the user and instance ID
AUTO_DBA_DOCKER_CONTAINER=$(docker ps | grep pgautodba-${USER}-${INSTANCE_ID} | awk '{print $1}')

# Check if the Docker container ID was found
if [ -z "$AUTO_DBA_DOCKER_CONTAINER" ]; then
    echo "No Docker container found for pgautodba-${USER}-${INSTANCE_ID}"
    exit 1
fi

# Execute backup commands inside the Docker container
docker exec -it $AUTO_DBA_DOCKER_CONTAINER sh -c 'cd .. && rm -rf backups && ./backup.sh --suffix recent && tar -czvf /home/autodba/backup.tar.gz /home/autodba/backups'

# Check if the backup command succeeded
if [ $? -ne 0 ]; then
    echo "Backup command failed inside the Docker container"
    exit 1
fi

# Create the backup directory if it doesn't exist
mkdir -p ${BACKUP_DIR}

# Copy the backup file from the Docker container to the host machine
docker cp $AUTO_DBA_DOCKER_CONTAINER:/home/autodba/backup.tar.gz ${BACKUP_DIR}/backup.tar.gz

# Check if the copy command succeeded
if [ $? -ne 0 ]; then
    echo "Failed to copy the backup file from the Docker container"
    exit 1
fi

# Print a success message
echo "Backup successful. File located at ${BACKUP_DIR}/backup.tar.gz"
