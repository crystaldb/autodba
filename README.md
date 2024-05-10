
# AutoDBA Project
This is an automated Database Administrator system for PostgreSQL databases.
The AutoDBA agent monitors and optimizes the database.

## Prerequisites
- Docker
- Docker Compose

## Structure
- `agent/`: Handles the main agent tasks like metrics collection, recommendations, and configuration.
- `api/`: Defines the API endpoints and services.
- `training/`: Handles machine learning models training.
- `gym/`: Includes simulation and benchmarking logic.
- `workloads/`: Deals with workload generation and management.

## Deployment
- `Dockerfile`: Defines the Docker setup for the agent.
- `docker-compose.yml`: to use a local Postgres DB, we use Docker Compose

## Setup Instructions

1. Clone the repository:

    ```bash
    git clone git@github.com:crystalcld/pgAutoDBA.git
    cd pgAutoDBA
    ```

2. Build the Docker images with Docker Compose:

    ```bash
    docker-compose build
    ```

3. Start the containers in the background:

    ```bash
    docker-compose up -d
    ```

4. Start the containers and rebuild if necessary:

    ```bash
    docker-compose up -d --build
    ```

5. Access the Agent's local PostgreSQL database directly via `psql`:

    ```bash
    docker-compose exec db psql --username=autodba_db_user --dbname=autodba_db
    ```

6. Inspect the Docker volume for PostgreSQL data:

    ```bash
    docker volume inspect pgautodba_postgres_data
    ```
