
# AutoDBA Scripts

This directory contains a set of scripts for managing the AutoDBA project. Below is a summary of the provided scripts and their usage instructions.

## Scripts

1. **build.sh**: This script is used to build and release AutoDBA in different formats, such as `.tar.gz`, `.deb`, or `.rpm` packages.

### Usage:
```
./build.sh [TARGET=all]
```
This will build the AutoDBA binaries for multiple architectures, prepare the web UI, include Prometheus, and generate release packages in the `build_output` directory.

The optional argument to this script can specify the target. The possible choices are:
 - `all` (default)
 - `tar.gz`
 - `rpm`
 - `deb`
 - `source`

2. **install.sh**: This script installs the AutoDBA project and its dependencies, including Prometheus. It supports both system-wide and user-specific installations.

### Usage:
```
./install.sh [--system] [--install-dir <DIRECTORY>] [--package <PACKAGE_FILE>] [--config <CONFIG_FILE>]
```
- `--system`: Install system-wide under `/usr/local/autodba`. This flag also installs autodba as a service.
- `--install-dir`: Specify a custom installation directory. If not specified, `$HOME/autodba` is used.
- `--package`: Provide a specific package file (e.g., `.tar.gz`, `.deb`, `.rpm`) for installation. If not provided, the script calls `build.sh tar.gz` to create the package.
- `--config`: Optionally provide a JSON configuration file that defines environment variables for AutoDBA.

The config file should have this format:
```json
{
    "DB_CONN_STRING": "<CONNECTION_STRING_TO_YOUR_POSTGRES_DB>",
    "AWS_ACCESS_KEY_ID": "<YOUR_AWS_ACCESS_KEY_ID>",
    "AWS_SECRET_ACCESS_KEY": "<YOUR_AWS_SECRET_ACCESS_KEY>",
    "AWS_REGION": "<YOUR_AWS_REGION>",
    "AWS_RDS_INSTANCE": "<YOUR_RDS_POSTGRES_INSTANCE_NAME>"
}
```

If a config file is not provided, the script will attempt to read from stdin. If neither is provided, the configuration will be generated from environment variables. Here are the exepected environment variables: `DB_CONN_STRING`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`, and `AWS_RDS_INSTANCE`.

If the input is provided via stdin, it will be validated as valid JSON before use. If invalid JSON is provided, the script will throw an error and exit.

3. **uninstall.sh**: This script uninstalls AutoDBA from the system, removing its binaries, Prometheus, and configuration files.

### Usage:
```
./uninstall.sh [--system] [--install-dir <DIRECTORY>]
```
- `--system`: Install system-wide under `/usr/local/autodba`. This flag also uninstalls autodba service.
- `--install-dir`: Specify a custom installation directory. If not specified, `$HOME/autodba` is used.

4. **run.sh**: This script is used to run AutoDBA in a Docker container. It sets up a Docker environment, runs the AutoDBA image, and exposes Prometheus and BFF ports.

### Usage:
```
./run.sh --db-url <TARGET_DATABASE_URL> [--instance-id <INSTANCE_ID>] [--rds-instance <RDS_INSTANCE_NAME>] [--disable-data-collection] [--continue]
```
- `--db-url`: Required parameter for the target database URL.
- `--instance-id`: Specify a unique instance ID if running multiple agents.
- `--rds-instance`: Collect metrics from an AWS RDS instance.
- `--disable-data-collection`: Disable data collection from the target database.

### Notes
- Use `sudo` for running system-wide `install.sh` and `uninstall.sh` scripts.
