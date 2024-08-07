# AutoDBA Project Private Repository

AutoDBA is the AI-powered PostgreSQL management agent developed at CrystalDB.
This repository contains the internal and proprietary codebase.

## Prerequisites

Your development environment should have:
- python 3.9+
- Docker
- Access to a Kubernetes cluster

## Structure
- *moving to public repo* `src/agent/`: Handles the main agent tasks like metrics collection, recommendations, and configuration.
- *moving to public repo* `src/api/`: Defines the API endpoints and services.
- `src/training/`: Handles machine learning models training.
- `gym/`: Includes simulation and benchmarking logic.

## Deployment
- *moving to public repo* `src/Dockerfile`: Defines the Docker setup for the agent.

## Setup Instructions

### Clone the repository:

    ```bash
    git clone git@github.com:crystalcld/pgAutoDBA.git
    ```

### Build and run the Gym

See [Gym documentation](gym/v2/README.md).

### Build and run the AutoDBA agent:

    ```bash
    cd pgAutoDBA
    docker build . -t autodba
    docker run --name pgautodba -e AUTODBA_TARGET_DB="<CONNECTION_STRING_TO_YOUR_TARGET_DB>" -e AWS_RDS_INSTANCE=<YOUR AWS DATABASE NAME> -e AWS_ACCESS_KEY_ID=<YOUR_AWS_ACCESS_KEY_ID> -e AWS_SECRET_ACCESS_KEY=<YOUR_AWS_SECRET_ACCESS_KEY> -e AWS_REGION=<YOUR_AWS_REGION> -p 8081:8080 -p 3001:3000 autodba
    ```

    The `AUTODBA_TARGET_DB` environment variable is necessary to connect AutoDBA to your target
    PostgreSQL database that is being managed by AutoDBA. You should assign the connection string
    to this environment variable, e.g., `postgresql://my_user:my_pass@localhost:5432/my_db`.

    If you include the `AWS_RDS_INSTANCE` argument, then the
    `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and `AWS_REGION` environment
    variables are all necessary to connect to the AWS RDS Instance target that
    is being monitored by AutoDBA.

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
