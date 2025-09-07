# ket - Kubernetes Embedded Testing

A Go-based tool for deploying and running tests in isolated Kubernetes environments.

## Architecture

```mermaid
sequenceDiagram
    participant User as Localhost/Dev Env
    participant Ket as ket Binary
    participant K8sAPI as Kubernetes API
    participant TestNamespace as Test Namespace
    participant TestJob as Test Runner Job
    participant TestPod as Test Runner Pod

    User->>Ket: ket launch --test-command "npm test"
    Ket->>Ket: Generate unique namespace name
    Ket->>K8sAPI: Create Namespace
    K8sAPI->>TestNamespace: Create namespace
    TestNamespace-->>K8sAPI: Namespace created
    K8sAPI-->>Ket: Namespace ready

    Ket->>K8sAPI: Create ServiceAccount
    K8sAPI->>TestNamespace: Deploy ServiceAccount
    TestNamespace-->>K8sAPI: ServiceAccount ready
    K8sAPI-->>Ket: ServiceAccount created

    Ket->>K8sAPI: Create Role (RBAC permissions)
    K8sAPI->>TestNamespace: Deploy Role
    TestNamespace-->>K8sAPI: Role ready
    K8sAPI-->>Ket: Role created

    Ket->>K8sAPI: Create RoleBinding
    K8sAPI->>TestNamespace: Deploy RoleBinding
    TestNamespace-->>K8sAPI: RoleBinding ready
    K8sAPI-->>Ket: RoleBinding created

    Ket->>K8sAPI: Create Job with mounted source code
    K8sAPI->>TestNamespace: Schedule Job
    TestNamespace->>TestJob: Create Job resource
    TestJob->>TestPod: Schedule Pod
    TestPod-->>TestJob: Pod running
    TestJob-->>K8sAPI: Job active
    K8sAPI-->>Ket: Job created

    Ket->>TestPod: Mount /workspace (source code)
    TestPod-->>Ket: Source code mounted

    Ket->>TestPod: Mount /reports (empty dir)
    TestPod-->>Ket: Reports directory ready

    Ket->>TestPod: Set environment variables
    Note over TestPod: KET_TEST_NAMESPACE<br/>KET_PROJECT_ROOT<br/>KET_WORKSPACE_PATH

    Ket->>TestPod: Execute test command
    TestPod->>TestPod: Run tests (e.g., npm test)
    TestPod->>TestNamespace: Create test resources (if needed)
    TestNamespace-->>TestPod: Test resources ready

    loop Test Execution
        TestPod->>TestPod: Execute test cases
        TestPod-->>Ket: Stream stdout/stderr
        Ket-->>User: Real-time test output
        TestPod->>TestPod: Write reports to /reports
    end

    TestPod-->>TestJob: Tests complete (success/failure)
    TestJob-->>K8sAPI: Job finished
    K8sAPI-->>Ket: Job completed

    Ket->>K8sAPI: Delete Namespace (cleanup)
    K8sAPI->>TestNamespace: Delete namespace
    TestNamespace-->>K8sAPI: Namespace deleted
    K8sAPI-->>Ket: Cleanup complete

    Ket-->>User: Test results & exit code
```

## Key Components

- **Isolated Testing**: Each test run gets a unique namespace
- **RBAC Setup**: ServiceAccount, Role, and RoleBinding for test permissions
- **Source Code Mounting**: HostPath volume for direct code access
- **Environment Variables**: Test scripts have access to namespace and path info
- **Automatic Cleanup**: Resources cleaned up after test completion

## Command Reference

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config, -c` | Path to config file | - |
| `--debug, -v` | Enable debug logging | `false` |
| `--image, -i` | Runner image | `node:18-alpine` |
| `--cluster-workspace-path, -w` | Workspace path in pod | `/workspace` |
| `--project-root, -r` | Project root path | `.` |

### Launch Flags

| Flag | Description | Default | Required |
|------|-------------|---------|----------|
| `--test-command, -t` | Test command to execute | - | ✅ |
| `--keep-namespace, -k` | Keep test namespace | `false` | ❌ |
| `--backoff-limit, -b` | Job backoff limit | `1` | ❌ |
| `--active-deadline-seconds, -d` | Job deadline in seconds | `1800` | ❌ |

### Commands

- `ket launch` - Run tests in Kubernetes
- `ket manifest` - Generate Kubernetes manifests
- `ket env` - Show environment variables documentation

## Development

### Project Structure

```
pkg/
├── config/     # Configuration and file loading
├── kube/       # Kubernetes operations
│   ├── apply/  # Cluster resource application
│   ├── generate/ # Kubernetes object generation
│   └── manifest/ # YAML marshaling
├── launcher/   # Job launch orchestration
└── logger/     # Structured logging

cmd/
└── testrunner/ # CLI entry point
```

### Building

```bash
make build      # Build binary
make test       # Run tests
make lint       # Lint code
make clean      # Clean artifacts
```

### Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage-summary

# Run specific package
make test-coverage-pkg PKG=pkg/launcher
```

### Requirements

- Go 1.24+
- Kubernetes cluster (Kind recommended)
- kubectl configured
- Docker (for building)
