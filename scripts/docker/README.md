# AutoDBA Docker Automation Scripts

This directory contains various scripts for automating tasks related to the Docker container running the AutoDBA agent.

## Scripts

### backup_script.sh

This script is used to backup the Prometheus and PostgreSQL databases inside the Docker container. It finds the Docker container based on the user and instance ID, executes the backup commands inside the container, and then copies the backup file to a specified directory on the host machine.

#### Usage

```bash
./backup_script.sh [--backup-dir <directory>] [--instance-id <id>]
```

#### Options

- `--backup-dir <directory>`: Specify the directory to save the backup (default: `./autodba_backups_dir`).
- `--instance-id <id>`: Specify the instance ID (default: `0`).

#### Example

```bash
./backup_script.sh --backup-dir /path/to/backup --instance-id 1
```

## Prerequisites

- Docker must be installed and running.
- The `pgautodba` Docker container should be running with the appropriate instance ID.
