# AutoDBA Agent Automation Scripts

This directory contains various scripts for automating tasks related to the AutoDBA agent.

## Scripts

### backup.sh

This script is used to create backups of the Prometheus and PostgreSQL databases managed by the AutoDBA agent.

#### Usage

The script takes an optional `--suffix` parameter to specify the suffix for backup files. If not provided, it defaults to the current timestamp.

```bash
./backup.sh [--suffix <suffix>]
```

#### Options

- `--suffix <suffix>`: Specify the suffix for backup files (default: current timestamp).

#### Environment Variables

- `POSTGRES_USER`: The PostgreSQL user.
- `POSTGRES_HOST`: The PostgreSQL host.
- `POSTGRES_PORT`: The PostgreSQL port.
- `POSTGRES_DB`: The PostgreSQL database name.
