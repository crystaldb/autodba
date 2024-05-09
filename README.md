
# AutoDBA Project
This is an automated Database Administrator system for PostgreSQL databases.
The AutoDBA agent monitors and optimizes the database.

## Structure
- `agent/`: Handles the main agent tasks like metrics collection, recommendations, and configuration.
- `api/`: Defines the API endpoints and services.
- `training/`: Handles machine learning models training.
- `gym/`: Includes simulation and benchmarking logic.
- `workloads/`: Deals with workload generation and management.

## Deployment
- `Dockerfile`: Defines the Docker setup for the agent.
- `docker-compose.yml`: to use a local Postgres DB, we use Docker Compose

