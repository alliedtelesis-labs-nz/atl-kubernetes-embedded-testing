# Kubernetes Test Runner

Run integration tests in isolated Kubernetes environments with zero setup.

## Quick Start

```bash
# Build and install
make build && make install

# Run tests
ket launch --test-command "npm test"
```

## Usage

### Basic Test Execution

```bash
# Run tests in current directory
ket launch --test-command "npm test"

# Generate manifests without running
ket manifest --test-command "npm test"
```

### Test Execution With Config File

Configuration can be provided in a JSON or YAML file, and will be automatically loaded from the current directory, or specified with the `--config` flag.

```bash
# Run tests with config file
ket launch --config ket-config.yaml
```

Example `ket-config.json` File:

```json
{
    "mode": "launch",
    "projectRoot": ".",
    "image": "atidyshirt/kubernetes-embedded-test-runner-node:latest",
    "testCommand": "npm run test:internal:mocha-integration",
    "clusterWorkspacePath": "/workspace",
    "debug": false,
    "logging": {
        "prefix": true,
        "timestamp": true
    }
}
```

### Environment Variables

Test scripts have access to:

- `KET_TEST_NAMESPACE` - The test namespace
- `KET_PROJECT_ROOT` - Project root path
- `KET_WORKSPACE_PATH` - Mounted workspace path

For more information on these environment variables, run the `ket env` command.

```bash
ket env
```

### Volume Mounts

- `/workspace` - (Required) Your source code - use `kind-config.yaml` or similar to mount this directory
- `/reports` - (Optional) Write test artifacts here

## Requirements

Runtime:

- Kubernetes cluster (Kind recommended)
    * Note, a `kind-config.yaml` or similar will be needed to setup mounting of source code
        * Examples can be found in the `example/` directory, consider how this maps to the `ket-config.yaml` file
- Docker image with dependencies for your Test Runner
    * e.g.: A `node:22.15` container with `npm`,`node` + `kubectl` installed
- Docker

Build:

- Go 1.24+

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Lint code
make clean    # Clean artifacts
```

## Architecture

For detailed architecture and technical documentation, see [`kubernetes-embedded-testing/README.md`](kubernetes-embedded-testing/README.md).
