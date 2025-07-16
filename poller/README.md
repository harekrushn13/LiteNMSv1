# Polling Engine - AugmentLNMS

The Polling Engine is a critical component of the AugmentLNMS system, responsible for collecting metrics from network devices at scheduled intervals. It implements an efficient, scalable architecture for device polling with prioritized scheduling and parallel execution.

## Architecture Overview

The Polling Engine follows a modular architecture with several key components:

```
┌─────────────────────────────────────────────────────────────┐
│                    Polling Engine                           │
│                                                             │
│  ┌─────────────┐         ┌─────────────┐                    │
│  │    ZMQ      │         │   Device    │                    │
│  │   Server    │────────►│  Manager    │                    │
│  └─────────────┘         └──────┬──────┘                    │
│         ▲                       │                           │
│         │                       ▼                           │
│         │                ┌─────────────┐                    │
│         │                │    Task     │                    │
│         │                │  Scheduler  │                    │
│         │                └──────┬──────┘                    │
│         │                       │                           │
│         │                       ▼                           │
│         │                ┌─────────────┐                    │
│  ┌─────────────┐         │   Worker    │                    │
│  │    Data     │◄────────┤    Pool     │────────┐           │
│  │  Collector  │         └─────────────┘        │           │
│  └─────────────┘                                │           │
│         │                                       │           │
│         │                                       ▼           │
│         │                               ┌─────────────┐     │
│         └───────────────────────────────┤     SSH     │     │
│                                         │   Client    │     │
│                                         └─────────────┘     │
└─────────────────────────────────────────────────────────────┘
                                                │
                                                ▼
                                         ┌─────────────┐
                                         │  Network    │
                                         │  Devices    │
                                         └─────────────┘
```

## Key Components

### ZMQ Server

- **Polling Server**: Handles communication with the Backend System
  - Receives provisioned device information via ZMQ PULL socket
  - Sends collected metrics data via ZMQ PUSH socket
  - Uses MessagePack for efficient serialization

### Device Manager

- **Device Registry**: Maintains a map of devices organized by counter ID
- **Provisioning Handler**: Processes new device information from the Backend
- **Task Generator**: Creates polling tasks for each counter and device combination

### Task Scheduler

- **Priority Queue**: Maintains a heap-based priority queue of polling tasks
- **Scheduler Loop**: Continuously checks for tasks that are due for execution
- **Task Dispatcher**: Sends due tasks to worker pool for execution

### Worker Pool

- **Parallel Workers**: Multiple workers for concurrent polling operations
- **Task Execution**: Processes polling tasks from the scheduler
- **Resource Management**: Limits concurrent connections to prevent overload

### SSH Client

- **Secure Connections**: Establishes SSH connections to network devices
- **Command Execution**: Runs specific commands based on counter ID
- **Data Parsing**: Processes command output into structured metrics

### Data Collector

- **Event Batching**: Collects individual polling events into batches
- **Batch Timing**: Sends batches at regular intervals
- **Data Formatting**: Ensures consistent data format for the Report Database

## Data Flow

1. **Device Provisioning Flow**:
   - Backend sends provisioned device information to ZMQ Server
   - ZMQ Server forwards device information to Device Manager
   - Device Manager updates device registry
   - Device Manager generates polling tasks for new devices
   - Tasks are added to the Priority Queue

2. **Polling Flow**:
   - Scheduler checks Priority Queue for due tasks
   - Due tasks are sent to Worker Pool
   - Workers establish SSH connections to devices
   - Commands are executed based on counter ID
   - Results are collected and formatted as Events
   - Events are batched by the Data Collector
   - Batches are sent to the Backend via ZMQ

## Polling Mechanism

The Polling Engine uses a sophisticated scheduling mechanism:

1. **Task Prioritization**: Tasks are ordered by next execution time
2. **Dynamic Scheduling**: Each counter can have a different polling interval
3. **Parallel Execution**: Multiple workers poll devices concurrently
4. **Resource Control**: Connection pool limits prevent device overload
5. **Error Handling**: Failed polls are logged but don't disrupt other operations

## Supported Counters

The Polling Engine supports various counter types defined in `config/counter.json`:

- **Memory Usage**: Free memory in bytes
- **CPU Utilization**: CPU usage percentage
- **Hostname**: Device hostname
- **Custom Counters**: Additional counters can be defined in the configuration

## Configuration

### Main Configuration

Configuration is stored in `config/config.json`:

```json
{
  "deviceBuffer": 10,
  "dataBuffer": 100,
  "workers": 5,
  "eventBuffer": 1000,
  "batchInterval": 5000
}
```

- `deviceBuffer`: Size of the device channel buffer
- `dataBuffer`: Size of the data channel buffer
- `workers`: Number of concurrent polling workers
- `eventBuffer`: Size of the event channel buffer
- `batchInterval`: Interval in milliseconds for sending batched events

### Counter Configuration

Counter definitions are stored in `config/counter.json`:

```json
{
  "1": {
    "name": "Memory Usage",
    "type": "uint64",
    "polling": 60
  },
  "2": {
    "name": "CPU Utilization",
    "type": "float64",
    "polling": 30
  },
  "3": {
    "name": "Hostname",
    "type": "string",
    "polling": 3600
  }
}
```

- `name`: Human-readable name of the counter
- `type`: Data type (uint64, float64, string)
- `polling`: Polling interval in seconds

## Building and Running

### Prerequisites

- Go 1.18 or higher
- ZeroMQ 4.3 or higher
- SSH client libraries

### Build Instructions

```bash
cd poller
go build
```

### Run Instructions

```bash
./poller
```

The Polling Engine will start and listen for device information from the Backend System.

## ZMQ Communication

### Incoming Messages (PULL Socket)

- **Socket**: tcp://*:6002
- **Format**: MessagePack-encoded array of Device objects
- **Content**: Device information including IP, credentials, and identifiers

### Outgoing Messages (PUSH Socket)

- **Socket**: tcp://*:6001
- **Format**: MessagePack-encoded array of Events objects
- **Content**: Collected metrics with object ID, counter ID, timestamp, and value

## Data Types

### Device

```go
type Device struct {
    ObjectID     uint32 `msgpack:"object_id" json:"object_id"`
    IP           string `msgpack:"ip" json:"ip"`
    CredentialID uint16 `msgpack:"credential_id" json:"credential_id"`
    DiscoveryID  uint16 `msgpack:"discovery_id" json:"discovery_id"`
    Username     string `msgpack:"username" json:"username"`
    Password     string `msgpack:"password" json:"password"`
    Port         uint16 `msgpack:"port" json:"port"`
}
```

### Events

```go
type Events struct {
    ObjectId  uint32      `msgpack:"objectId" json:"objectId"`
    CounterId uint16      `msgpack:"counterId" json:"counterId"`
    Timestamp uint32      `msgpack:"timestamp" json:"timestamp"`
    Value     interface{} `msgpack:"value" json:"value"`
}
```

### Task

```go
type Task struct {
    CounterID      uint16
    NextExecution  time.Time
    Interval       time.Duration
    Index          int
}
```

## Error Handling

The Polling Engine implements comprehensive error handling:

1. **Connection Errors**: Failed SSH connections are logged and retried on next schedule
2. **Command Errors**: Failed commands are logged without disrupting other operations
3. **ZMQ Errors**: Communication errors are logged and operations continue when possible
4. **Graceful Shutdown**: Clean shutdown process ensures all resources are released

## Logging

The Polling Engine uses the zap logging library for structured logging:

- **Info**: Normal operation information
- **Warn**: Potential issues that don't affect operation
- **Error**: Errors that affect operation but don't cause shutdown
- **Debug**: Detailed information for troubleshooting (when enabled)

## Security Considerations

- **SSH Credentials**: Device credentials are stored in memory only
- **Secure Connections**: SSH is used for secure device communication
- **Minimal Permissions**: Device accounts should have minimal required permissions
- **Connection Limits**: Resource pools prevent connection flooding

## Project Structure

```
poller/
├── config/
│   ├── config.json         # Main configuration
│   └── counter.json        # Counter definitions
├── logger/
│   └── logger.go           # Logging configuration
├── polling/
│   ├── poller.go           # Main polling logic
│   ├── provision.go        # Device provisioning
│   ├── queue.go            # Priority queue implementation
│   └── scheduler.go        # Task scheduling and execution
├── server/
│   └── server.go           # ZMQ server implementation
├── utils/
│   ├── config.go           # Configuration loading
│   └── types.go            # Common type definitions
├── main.go                 # Application entry point
└── README.md               # This documentation
```

## Future Improvements

- **Credential Encryption**: Encrypt device credentials in memory
- **Plugin System**: Support for custom polling methods
- **Adaptive Polling**: Dynamically adjust polling intervals based on device load
- **SNMP Support**: Add SNMP polling capabilities
- **Metrics Collection**: Add internal metrics for monitoring polling performance
- **High Availability**: Support for clustered operation
