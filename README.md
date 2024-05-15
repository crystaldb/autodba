
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

2. Build and run the project:

    ```bash
    ./run.sh
    ```

    Note 1: The database is created automatically if it doesn't exist. If you apply any schema changes (to `src/agent/init/schema.sql`), you need to manually recreate the database, as currently there's no auto-upgrade mechanism in place:
    ```bash
    ./run.sh --recreate
    ```
    Note 2: For production, you need to add a `.env.prod` file that contains the required environment variable, and add the `--env prod` argument:
    ```bash
    ./run.sh --env prod # and optionally pass --recreate to recreate the database
    ```

3. Seed the database using the provided `seed_db` command:

    ```bash
    docker exec --env-file .env.dev -it pgautodba python manage.py seed_db
    ```

4. Run the tests (via a running container using `Step 2`):

    ```bash
    docker exec --env-file .env.dev -it pgautodba bash -c "cd /home && pytest"
    ```

5. View all logs (which are under `/home/src/logs` on the docker container):

    ```bash
    docker exec --env-file .env.dev -it pgautodba /home/scripts/view-logs.sh
    ```

6. Access the Agent's local PostgreSQL database directly via `psql`:

    ```bash
    docker exec -it pgautodba psql --username=autodba_db_user --dbname=autodba_db
    ```

7. Inspect the Docker volume for PostgreSQL data:

    ```bash
    docker volume inspect autodba_postgres_data
    ```
