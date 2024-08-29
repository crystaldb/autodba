# AutoDBA for PostgreSQL

AutoDBA is an AI agent that helps manage PostgreSQL databases.
The project aims to ensure that everyone who runs PostgreSQL can have a database expert at their side at all times.

## Motivation

We all want our production PostgreSQL databases to run well—they should be reliable, performant, efficient, scalable, and secure.
If your team is fully staffed with skilled database administrators, then you are in luck.
If not, you may find yourself working reactively, dealing with problems as they arise and looking up whatever you need as you go.
If you are a software engineer or a site reliability engineer who is responsible for your organization's databases, you probably have many other things to do that would be a better use of your time than solving database problems.

## Project Status: Observability Only

This project presently includes only the basic observability features of AutoDBA.
Building an AI agent requires quality data, so we first need to make sure that this foundation is solid.

If it is not possible for a human expert to look at the data, understand what is going on, and make a good decision, then it is probably asking too much of the AI agent to expect it to do so.

The core AI agent is under development and will be released once it reaches a reasonable degree of accuracy and stability.

Though this release is limited, the observability features provided here are a contribution beyond what is otherwise available as open source.


## Limitations

This is an early release of AutoDBA so you will find many wished for features missing.

Temporary limitations:

- Only compatiable with PostgreSQL version 16.
- Only works with AWS RDS PostgreSQL.


## Roadmap

Some of the near-term items on our roadmap include the following:

- Observability
    - [x] Database and system metrics
    - [x] Wait events
    - [ ] Health dashboard
    - [ ] Query normalization + PII filtering
    - [ ] Log analysis
    - [ ] Fleet overview
- Alerting
    - [ ] Accurate (low noise) alerting on current status
    - [ ] Predictive health alerts


## Contributing

We welcome contributions to AutoDBA! Contributor guidelines are under development so [please reach out](mailto:jssmith@crystal.cloud) if you are interested in working with us.


## About the Authors

AutoDBA is developed the engineers and database experts at  [CrystalDB](https://www.crystaldb.cloud/).
Our mission is to make it easy for you to run your database well, so you can focus on building better software.
CrystalDB also offers commercial support for AutoDBA and PostgreSQL.



# Installation

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
    git clone git@github.com:crystaldb/autodba.git
    ```

### Build and run the Gym

See [Gym documentation](gym/v2/README.md).

### Build and run the AutoDBA agent:

    ```bash
    cd autodba
    docker build . -t autodba
    docker run --name autodba -e AUTODBA_TARGET_DB="<CONNECTION_STRING_TO_YOUR_TARGET_DB>" -e AWS_RDS_INSTANCE=<YOUR AWS DATABASE NAME> -e AWS_ACCESS_KEY_ID=<YOUR_AWS_ACCESS_KEY_ID> -e AWS_SECRET_ACCESS_KEY=<YOUR_AWS_SECRET_ACCESS_KEY> -e AWS_REGION=<YOUR_AWS_REGION> -p 8081:8080 -p 3001:3000 autodba
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
    cd autodba
    docker build . --target test
    docker build . --target lint
    ```

4. Access the Agent's local PostgreSQL database directly via `psql`:

    ```bash
    docker exec -it autodba psql --username=autodba_db_user --dbname=autodba_db
    ```
    TODO: Setup environment variables so psql doesn't need those CLI arguments

5. Setup + inspect a Docker volume for PostgreSQL data:
    TODO: Write directions explaining how to use a volume or bind mount to preserve the database for debugging.  Then (for volumes):
    ```bash
    docker volume inspect autodba_postgres_data
    ```
