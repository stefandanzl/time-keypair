# Taskfile Documentation

This project uses [Task](https://taskfile.dev/) as a task runner and build tool. Task is a simpler alternative to Make, with a focus on helping you get things done quickly and easily.

## Installation

If you haven't installed Task yet, visit [taskfile.dev/#/installation](https://taskfile.dev/#/installation) for instructions.

Quick installation methods:

- **macOS**: `brew install go-task/tap/go-task`
- **Linux**: `sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin`
- **Windows**: `choco install go-task` or `scoop install task`

## Available Tasks

You can list all available tasks with:

```bash
task --list
```

Here are the main tasks defined:

### Development Tasks

- `task build`: Build the application
- `task run`: Run the application
- `task test`: Run tests
- `task clean`: Clean build artifacts

### Docker Tasks

- `task docker-build`: Build Docker image
- `task docker-run`: Run Docker container
- `task docker-compose-up`: Run with Docker Compose
- `task docker-compose-down`: Stop Docker Compose services

## Environment Variables

All tasks accept environment variables. You can override the defaults by:

1. Setting environment variables:

```bash
PORT=9090 SUPER_ADMIN_KEY=my_custom_key task run
```

2. Creating a `.env` file in the project root.

The following environment variables are supported:

- `PORT`: HTTP server port (default: 8080)
- `SUPER_ADMIN_KEY`: Super admin key for administration (default: "super_admin_key")
- `CONFIG_FILE_PATH`: Path to the configuration file (default: "./config/config.json")
- `AUTO_SAVE_INTERVAL`: Interval in seconds for auto-saving (default: 60)

## Examples

Running the server with custom settings:

```bash
PORT=9090 SUPER_ADMIN_KEY=secure_key task run
```

Running Docker with custom settings:

```bash
PORT=9090 SUPER_ADMIN_KEY=secure_key task docker-run
```
