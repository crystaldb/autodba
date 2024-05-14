
# AutoDBA Project
This is an automated Database Administrator system for PostgreSQL databases.
The AutoDBA agent monitors and optimizes the database.

## Prerequisites
- Docker

## Structure
- `agent/`: Handles the main agent tasks like metrics collection, recommendations, and configuration.
- `api/`: Defines the API endpoints and services.
- `training/`: Handles machine learning models training.
- `gym/`: Includes simulation and benchmarking logic.
- `workloads/`: Deals with workload generation and management.

## Deployment
- `Dockerfile`: Defines the Docker setup for the agent.

## Setup Instructions

1. Clone the repository:

    ```bash
    git clone git@github.com:crystalcld/pgAutoDBA.git
    cd pgAutoDBA
    ```

2. Build and run the project (in development mode):

    ```bash
    ./run.sh
    ```

3. Build and run the project (in production mode):

    ```bash
    ./run.sh --env prod
    ```

    Note: you need to add a `.env.prod` file that contains the required environment variable.
    Note: the production database is not created automatically. You can create it using the initialization SQL script in `src/agent/init/schema.sql`:
    ```bash
    docker exec -it pgautodba /bin/bash -c "export FLASK_APP=\"api/endpoints.py\" && export FLASK_DEBUG=True && python manage.py create_db"
    ```

4. Seed the database using the provided `seed_db` command:

    ```bash
    docker exec -it pgautodba /bin/bash -c "export FLASK_APP=\"api/endpoints.py\" && export FLASK_DEBUG=True && python manage.py seed_db"
    ```

5. Access the Agent's local PostgreSQL database directly via `psql`:

    ```bash
    docker exec -it pgautodba psql --username=autodba_db_user --dbname=autodba_db
    ```

6. Inspect the Docker volume for PostgreSQL data:

    ```bash
    docker volume inspect autodba_postgres_data
    ```
