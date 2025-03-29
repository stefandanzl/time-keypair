# Multi-User Cron Server with Data Store

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
- `GET /cron/{user_key}/jobs`: List all jobs for a user
- `POST /cron/{user_key}/jobs`: Create a new job for a user
- `PUT /cron/{user_key}/jobs`: Update all jobs for a user
- `GET /cron/{user_key}/job/{job_id}`: Get a specific job
- `PUT /cron/{user_key}/job/{job_id}`: Update a specific job
- `DELETE /cron/{user_key}/job/{job_id}`: Delete a specific job

### Data Endpoints

- `GET /data/{user_key}/keys`: List all data keys for a user
- `GET /data/{user_key}/{data_key}`: Get data for a user
- `PUT /data/{user_key}/{data_key}`: Set data for a user
- `DELETE /data/{user_key}/{data_key}`: Delete data for a user

### Other Endpoints

- `GET /health`: Health check endpoint

## Getting Started

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

### Basic Functionality Test

Tests core API operations including user management, cron jobs, and data storage:

```bash
chmod +x ./test-api.sh
./test-api.sh
```

### Load Test

Tests performance under load by creating multiple users, jobs, and data entries concurrently:

```bash
chmod +x ./load-test.sh
./load-test.sh
```

### Error Handling Test

Tests how the API handles various error conditions and invalid inputs:

```bash
chmod +x ./error-test.sh
./error-test.sh
```

### Unit Tests

Runs all Go unit tests in the project:

```bash
task test
```

## Usage Examples

### Create a new user
```bash
curl -X POST http://localhost:8080/admin/super_admin_key/users -d '{"user":"user1"}'
```

### Create a new cron job
```bash
curl -X POST http://localhost:8080/cron/user1/jobs -d '{"id":"job1","cron":"0 * * * * *","url":"https://example.com","active":true}'
```

### Store data
```bash
curl -X PUT http://localhost:8080/data/user1/settings -d '{"theme":"dark","notifications":true}'
```

### Retrieve data
```bash
curl http://localhost:8080/data/user1/settings
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
