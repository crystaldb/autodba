# AutoDBA for PostgreSQL

AutoDBA is an AI agent that helps manage PostgreSQL databases.
This project aims to ensure that everyone who runs PostgreSQL has access to a skilled virtual database administrator at all times.

## Motivation

We all want our production PostgreSQL databases to run well—they should be reliable, performant, efficient, scalable, and secure.
If your team is fully staffed with database administrators, then you are in luck.
If not, you may find yourself working reactively, dealing with problems as they arise and looking up what to do as you go.

Oftentimes, operational responsibility for databases falls to software engineers site reliability engineers, who usually have many other things to do.
They have better ways to spend their time than tuning or troubleshooting databases.

Reliability and security are the top priorities for database operations, and our approach reflects that.
The success of autonomous vehicles shows that with careful engineering, one can build systems that operate safely in complex environments.
The time has come to do the same for similarly challenging IT operations—as AutoDBA does for databases.


## Project Status: Observability Only

As of today, this project includes only the basic observability features of AutoDBA.
Building an AI agent requires quality data, so we first need to make sure that this foundation is solid.

Why did we put effort into tools for visualizing and exploring data when we envision a future were only machines consume it?
For one, we need to verify that the data is sufficient to support good decision making—if a human expert does not have enough information to make a good decision, then we are probably asking too much of the AI agent if we expect it to do so.
Furthermore, AutoDBA's observability tools add value to the PostgreSQL ecosystem, even without an AI agent.

The core AI agent is under development and will be released once it reaches a reasonable degree of accuracy and stability.


## Temporary Limitations

This is an early release of AutoDBA's observability features.
We are committed to supporting PostgreSQL in all environments and poplular major versions.
However, the following *temporary limitations* are presently in place:

- Only compatiable with PostgreSQL version 16.
- Only works with AWS RDS PostgreSQL.


## Roadmap

Our near-term roadmap includes the following:

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

AutoDBA is developed by the engineers and database experts at  [CrystalDB](https://www.crystaldb.cloud/).
Our mission is to make it easy for you to run your database well, so you can focus on building better software.
CrystalDB also offers commercial support for AutoDBA and PostgreSQL.


## Frequently Asked Questions

### What is AutoDBA?

AutoDBA is an AI agent for operating PostgreSQL databases.
This means that it connects to an existing PostgreSQL database and takes actions, as necessary, to ensure that remains reliable, efficient, scalabe, and secure.


### Will AutoDBA replace my DBA?

Time will tell whether AI agents completely replace human database administrators (DBAs).
Our work suggests that AI agents will do some tasks much better than humans.
They can find patterns across large amounts of data, they are always available, and they respond instantly.
They can also draw upon extensive knowledge bases and operational datasets, allowing them to proceed with less trial and error than people.

On the other hand, people working in the team will have a more nuanced understanding of the needs of the business.
They will be better able to make high-level design decisions and to analyze trade-offs that impact the development process.


### Do I still need to hire a DBA if I use AutoDBA?

If you do not already have a DBA on staff, then chances are good that AutoDBA can allow you postpone hiring one, particularly if you have platform engineers or site reliability engineers who are interested in applying its recommendations.
CrystalDB and others also offer commercial support for PostgreSQL.


### Which PostgreSQL versions are supported?

Currently, AutoDBA is only compatible with PostgreSQL version 16.
We are working on expanding support to other versions.
Please share your thoughts on how far back we should go.


### Can I use AutoDBA with my on-premises PostgreSQL installation?

At present, AutoDBA only works with AWS RDS PostgreSQL.
Support for on-premises installations, Google Cloud SQL, and Azure SQL is coming.
We want AutoDBA to run anywhere that PostgreSQL runs.


### Will AutoDBA support databases other than PostgreSQL?

At present, we are fully focused on AutoDBA for PostgreSQL.
We expect to maintain that focus for the forseeable future.


### Is AutoDBA open source?

We believe that every PostgreSQL database should be managed by an AI agent.
In pursuit of this vision, we are releasing the core operational features of AutoDBA under the Apache 2.0 open source license.

For avoidance of doubt, AutoDBA is commercial open source software.
Certain enterprise features will be available only in commercial versions of the product.

As of this writing (August 2024), there is active debate what the term “open source” means for AI models.
Is a model open if the developer releases the weights but not the training data and methods?

We are committed to providing open weights and some training data.
However, we also expect to release models trained on proprietary data sets.


### What happens to the data that AutoDBA collects?

AutoDBA collects and stores operational metrics collected from your database.
It does not transmit this data to anyone.
You should keep the web interface secure to avoid exposing this data to others.


### How can I get support for AutoDBA?

CrystalDB offers commercial support for AutoDBA and PostgreSQL.
For more information or to discuss your needs, please contact us at [support@crystal.cloud](mailto:support@crystal.cloud).


### How can I support the AutoDBA project?

Foremost, use AutoDBA and give us feedback!

We also welcome feature suggestions, bug reports, or contributions to the codebase.


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
