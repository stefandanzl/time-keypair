# Time-Keypair - Multi-User Cron Server with Data Store

A web server that combines scheduled job execution with a data store system, supporting multiple users with isolated environments through API key authentication.

## Features

- Execute scheduled HTTP GET requests (cron jobs)
- Store and retrieve arbitrary data for each user
- Authenticate using keys in URL paths
- Thread-safe concurrent access
- Periodic auto-saving of configuration
- Docker container support with health check
- Configuration via environment variables

## Configuration

The server can be configured using environment variables:

- `PORT`: Port number (default: 8080)
- `SUPER_ADMIN_KEY`: Super admin key (default: "super_admin_key")
- `CONFIG_FILE_PATH`: Configuration file path (default: "/config/config.json")
- `AUTO_SAVE_INTERVAL`: Auto-save interval in seconds (default: 60)

## API Endpoints

### Admin Endpoints

- `GET /admin/{super_key}/users`: List all users
- `POST /admin/{super_key}/users`: Create a new user
- `DELETE /admin/{super_key}/users/{user}`: Delete a user
- `GET /admin/{super_key}/config`: Get full configuration
- `PUT /admin/{super_key}/config`: Replace full configuration

### Cron Endpoints

- `GET /status/{user_key}`: Get job statuses for a user
- `GET /cron/{user_key}`: List all jobs for a user
- `POST /cron/{user_key}`: Create a new job for a user
- `PUT /cron/{user_key}`: Update all jobs for a user
- `GET /cron/{user_key}/{job_id}`: Get a specific job
- `PUT /cron/{user_key}/{job_id}`: Update a specific job
- `DELETE /cron/{user_key}/{job_id}`: Delete a specific job
- `GET /cron/{user_key}/{job_id}/on`: Activate a specific job
- `GET /cron/{user_key}/{job_id}/off`: Deactivate a specific job

### Data Endpoints

- `GET /data/{user_key}/keys`: List all data keys for a user
- `GET /data/{user_key}/{data_key}`: Get data for a user
- `PUT /data/{user_key}/{data_key}`: Set data for a user
- `DELETE /data/{user_key}/{data_key}`: Delete data for a user

### Other Endpoints

- `GET /health`: Health check endpoint

## Getting Started

### Docker Run

#### Basic
````shell
docker run -p 8080:8080 \
        -e SUPER_ADMIN_KEY=super_admin_key \
        -v ./config:/config \
        ghcr.io/stefandanzl/time-keypair:latest
````

#### Detailed
````shell
docker run -p 8080:8080 \
        -e PORT=8080 \
        -e SUPER_ADMIN_KEY=super_admin_key \
        -e CONFIG_FILE_PATH=/config/config.json \
        -e AUTO_SAVE_INTERVAL=60 \
        -v ./config:/config \
        ghcr.io/stefandanzl/time-keypair:latest
````

### Prerequisites

- [Go](https://golang.org/dl/) version 1.20 or higher
- [Docker](https://www.docker.com/products/docker-desktop) (optional)
- [Task](https://taskfile.dev/#/installation) - a task runner / simpler Make alternative
- [jq](https://stedolan.github.io/jq/download/) - for parsing JSON in test scripts

### Build and Run

1. **Build and Run Locally**:
   ```bash
   task build
   task run
   ```

2. **Run with Docker**:
   ```bash
   task docker-run
   ```
   
3. **Run with Docker Compose**:
   ```bash
   task docker-compose-up
   ```

## Testing

The project includes several test scripts to validate functionality:

### Run All Tests

To run all tests (unit tests, API tests, load tests, and error handling tests):

```bash
task test
```

or

```bash
task test-all
```

### Individual Test Categories

- **Unit Tests**: `task test-unit`
- **API Functionality Tests**: `task test-api`
- **Load Tests**: `task test-load`
- **Error Handling Tests**: `task test-error`

### Test Scripts

All test scripts are located in the `tests` directory and can also be run directly:

```bash
bash ./tests/test-api.sh
bash ./tests/load-test.sh
bash ./tests/error-test.sh
bash ./tests/run-all-tests.sh
```

## Usage Examples

### Create a new user
```bash
curl -X POST http://localhost:8080/admin/super_admin_key/users -d '{"user":"user1"}'
```

### Create a new cron job
```bash
curl -X POST http://localhost:8080/cron/user1 -d '{"id":"job1","cron":"0 * * * * *","url":"https://example.com","active":true}'
```

### Store data
```bash
curl -X PUT http://localhost:8080/data/user1/settings -d '{"theme":"dark","notifications":true}'
```

### Retrieve data
```bash
curl http://localhost:8080/data/user1/settings
```

### Activate or deactivate a cron job
```bash
# Activate a job
curl http://localhost:8080/cron/user1/job1/on

# Deactivate a job
curl http://localhost:8080/cron/user1/job1/off
```

## Configuration Format

The server uses a single JSON file for all configuration and data:

```json
{
  "user1": {
    "cron": [
      {"id": "job1", "cron": "0 * * * * *", "url": "https://example.com", "active": true}
    ],
    "data": {
      "key1": "value1",
      "settings": {"theme": "dark", "notifications": true}
    }
  },
  "user2": {
    ...
  }
}
```

## Note on Cron Expressions

This server uses the [robfig/cron/v3](https://github.com/robfig/cron) package, which requires cron expressions to include a seconds field as the first value. For example:

- Standard cron: `* * * * *` (minute, hour, day of month, month, day of week)
- With this server: `0 * * * * *` (seconds, minute, hour, day of month, month, day of week)

The first value specifies seconds (0-59), followed by the standard cron fields.
