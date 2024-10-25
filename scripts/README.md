
# AutoDBA Scripts

This directory contains a set of scripts for managing the AutoDBA project. Below is a summary of the provided scripts and their usage instructions.

## Scripts

1. **build.sh**: This script is used to build and release AutoDBA.

### Usage:
```
./build.sh
```
This will build the AutoDBA binaries for multiple architectures, prepare the web UI, include Prometheus, and generate release packages in the `build_output` directory.

2. **install.sh**: This script installs the AutoDBA project and its dependencies, including Prometheus. It supports both system-wide and user-specific installations.

### Usage:
```
./install.sh [--system] [--install-dir <DIRECTORY>] [--config <CONFIG_FILE>]
```
- `--system`: Install system-wide under `/usr/local/autodba`. This flag also installs autodba as a service.
- `--install-dir`: Specify a custom installation directory. If not specified, `$HOME/autodba` is used.
- `--config`: Optionally provide an autodba configuration file that defines parameters for AutoDBA.

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
./run.sh --config <CONFIG_FILE> [--instance-id <INSTANCE_ID>]
```
- `--config`: Required parameter specifying the path to the configuration file.
- `--instance-id`: Specify a unique instance ID if running multiple instances of the agent.
```

### Notes
- Use `sudo` for running system-wide `install.sh` and `uninstall.sh` scripts.
