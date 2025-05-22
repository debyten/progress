# Progress Tracking Library

A robust Go library for managing and monitoring long-running operations through WebSocket connections. This library provides real-time progress updates and state management capabilities.

## Features

- Real-time progress tracking via WebSocket connections
- Thread-safe progress management
- Context-based progress injection
- Multiple client support per progress instance
- Automatic client connection management
- No-op implementation for cases where progress tracking is not needed

## Core Components

### Progress Interface

The core `Progress` interface defines the contract for progress tracking:

```go
type Progress interface {
    GetID() string
    StartSignal() chan struct{}
    Update(state string, details ...map[string]any)
}
```

### Progress Manager Service

The `Service` interface provides methods for managing progress instances:

```go
type Service interface {
    CreateProgress() Progress
    GetProgress(id string) (*progress, bool)
    DeleteProgress(id string)
    Stream(ctx context.Context, ws *websocket.Conn, id string)
}
```

## Usage

### Creating a Progress Instance

```go
// Initialize the service
service := progress.NewService()

// Create a new progress instance
prog := service.CreateProgress()
```

### Injecting Progress into Context

```go
ctx := progress.Inject(context.Background(), prog)
```

### Retrieving Progress from Context

```go
prog := progress.FromContext(ctx)
```

**Note: A no-operation (noop) progress implementation is returned from the context if no progress instance was
previously injected**

### Streaming Progress

The Server instance exposes a WebSocket endpoint at `/api/v1/progress/{id}` where `id` is the progress instance
identifier. Clients can connect to this endpoint to receive real-time progress updates for a specific operation. The
connection will automatically be managed by the server, sending updates whenever the progress state changes.

## WebSocket Message Format

Progress updates are sent to clients in the following JSON format:

```json
{
    "state": "Processing",
    "details": {
        "percentComplete": 50,
        "currentStep": "Validating data"
    }
}
```

## Thread Safety

The library implements thread-safe operations using mutex locks for all critical sections, ensuring safe concurrent access to progress instances.

## Best Practices

1. Always use the context injection mechanism for passing progress instances
2. Close WebSocket connections when they are no longer needed
3. Use the StartSignal channel when you need to ensure a client is connected before proceeding
4. Clean up progress instances using DeleteProgress when they are no longer needed

## Dependencies

- github.com/gorilla/websocket
- github.com/google/uuid