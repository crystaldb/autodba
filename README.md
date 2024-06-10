
# AutoDBA Project
This is an automated Database Administrator system for PostgreSQL databases.
The AutoDBA agent monitors and optimizes the database.

## Prerequisites
- Docker for agent development
- The gym requires Kubernetes (see the `gym` directory for more information)

## Structure
- `src/agent/`: Handles the main agent tasks like metrics collection, recommendations, and configuration.
- `src/api/`: Defines the API endpoints and services.
- `src/training/`: Handles machine learning models training.
- `gym/`: Includes simulation and benchmarking logic.

## Deployment
- `src/Dockerfile`: Defines the Docker setup for the agent.

## Setup Instructions

1. Clone the repository:

    ```bash
    git clone git@github.com:crystalcld/pgAutoDBA.git
    cd pgAutoDBA
    ```

2. Build and run the project:

    ```bash
    cd pgAutoDBA
    docker build . -t autodba
    docker run --name pgautodba -e AUTODBA_TARGET_DB="<CONNECTION_STRING_TO_YOUR_TARGET_DB>" -p 8081:8080 -p 3001:3000 autodba
    ```

    The `AUTODBA_TARGET_DB` environment variable is necessary to connect AutoDBA to your target
    PostgreSQL database that is being managed by AutoDBA. You should assign the connection string
    to this environment variable, e.g., `postgresql://my_user:my_pass@localhost:5432/my_db`.
    
    The `--name` option is optional.  For multi-user docker environments, make sure it's unique.
    
    Similarly, replace `8081` and `3001` with whatever port numbers should be bound on the Docker host for the AutoDBA Agent UI and the Grafana interface, respectively.

    Note: The agent's ephemeral private database is automatically created at startup.

    Note: You need to make sure that `pg_stat_statements` is installed on your target Postgres database:
    ```
    psql -c 'create extension if not exists pg_stat_statements;' '<CONNECTION_STRING_TO_YOUR_TARGET_DB>'
    ```

3. Run the unit tests + linter:

    ```bash
    cd pgAutoDBA
    docker build . --target test
    docker build . --target lint
    ```

4. Access the Agent's local PostgreSQL database directly via `psql`:

    ```bash
    docker exec -it pgautodba psql --username=autodba_db_user --dbname=autodba_db
    ```
    TODO: Setup environment variables so psql doesn't need those CLI arguments

5. Setup + inspect a Docker volume for PostgreSQL data:
    TODO: Write directions explaining how to use a volume or bind mount to preserve the database for debugging.  Then (for volumes):
    ```bash
    docker volume inspect autodba_postgres_data
    ```
