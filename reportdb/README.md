# Report Database (@reportdb) - AugmentLNMS

The Report Database (@reportdb) is a specialized time-series database designed for efficient storage and retrieval of network metrics data. It provides a high-performance, scalable solution optimized for write-heavy workloads with efficient query capabilities.

## Architecture Overview

The @reportdb follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                   Communication Layer                       │
│  ┌─────────────┐                     ┌─────────────┐        │
│  │   Polling   │                     │    Query    │        │
│  │   Server    │                     │    Server   │        │
│  └──────┬──────┘                     └──────┬──────┘        │
└─────────┼──────────────────────────────────┼────────────────┘
          │                                  │
┌─────────▼──────────────────────────────────▼────────────────┐
│                   Distribution Layer                        │
│  ┌─────────────┐                     ┌─────────────┐        │
│  │    Writer   │                     │    Reader   │        │
│  │    Broker   │                     │    Broker   │        │
│  └──────┬──────┘                     └──────┬──────┘        │
└─────────┼──────────────────────────────────┼────────────────┘
          │                                  │
┌─────────▼──────────────────────────────────▼────────────────┐
│                    Processing Layer                         │
│  ┌─────────────┐                     ┌─────────────┐        │
│  │   Writers   │                     │   Readers   │        │
│  └──────┬──────┘                     └──────┬──────┘        │
└─────────┼──────────────────────────────────┼────────────────┘
          │                                  │
┌─────────▼──────────────────────────────────▼────────────────┐
│                      Storage Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │  Store Pool │  │    Store    │  │ Data Fetcher│          │
│  └──────┬──────┘  │   Engine    │  └──────┬──────┘          │
│         │         └──────┬──────┘         │                 │
│         └──────────┬─────┴────────────────┘                 │
│                    │                                        │
│  ┌─────────────────▼───────────────────┐                    │
│  │           File Storage              │                    │
│  │  ┌─────────────┐  ┌─────────────┐   │                    │
│  │  │     File    │  │    Index    │   │                    │
│  │  │   Manager   │  │   Manager   │   │                    │
│  │  └─────────────┘  └─────────────┘   │                    │
│  └───────────────────────────────────┬─┘                    │
└──────────────────────────────────────┼────────────────────┬─┘
                                       │                    │
┌──────────────────────────────────────▼────────────────────▼─┐
│                      File System                            │
└─────────────────────────────────────────────────────────────┘
```

## Key Components

### Communication Layer

- **Polling Server**: Receives metrics data from the Backend via ZMQ
- **Query Server**: Handles query requests and responses via ZMQ

### Distribution Layer

- **Writer Broker**: Distributes incoming events to multiple writers
- **Reader Broker**: Distributes incoming queries to multiple readers

### Processing Layer

- **Writers**: Process and store incoming metrics data
- **Readers**: Process queries and retrieve data
- **Query Processor**: Executes query logic and aggregations

### Storage Layer

- **Store Pool**: Manages storage engine instances
- **Storage Engine**: Coordinates file and index operations
- **Data Fetcher**: Retrieves data from storage engines
- **File Manager**: Handles file operations and memory mapping
- **Index Manager**: Maintains indexes for efficient data retrieval

## Data Flow

### Write Path

```
Backend → Polling Server → Data Channel → Writer Broker → Writers → 
Store Pool → Storage Engine → File Manager/Index Manager → Disk
```

### Read Path

```
Backend → Query Server → Query Channel → Reader Broker → Readers → 
Data Fetcher → Store Pool → Storage Engine → File Manager/Index Manager → 
Query Processor → Response Channel → Query Server → Backend
```

## Storage Format

### File Organization

Data is organized hierarchically by date and counter ID:

```
/database/YYYY/MM/DD/counter_N/
```

For example:
```
/database/2023/05/15/counter_1/
```

### Data Format

Data is stored in binary format for efficiency:

```
[length(4)][timestamp(4)][value(N)]
```

Where:
- `length`: 4-byte integer representing the length of the value
- `timestamp`: 4-byte integer representing the Unix timestamp
- `value`: Variable-length data based on the counter type

### Partitioning

Data is partitioned based on object ID to improve parallel access:

- Each day/counter combination has multiple partition files
- Object IDs are hashed to determine the partition
- Each partition is a separate file with its own index

## Query Capabilities

The @reportdb supports several query types:

### Gauge Queries

Return a single aggregated value for a time range:

```json
{
  "counter_id": 1,
  "object_ids": [1001, 1002],
  "from": 1620000000,
  "to": 1620086400,
  "aggregation": "avg"
}
```

### Histogram Queries

Return time-series data with a specified interval:

```json
{
  "counter_id": 1,
  "object_ids": [1001, 1002],
  "from": 1620000000,
  "to": 1620086400,
  "interval": 3600,
  "aggregation": "avg",
  "group_by_objects": false
}
```

### Grid Queries

Return data grouped by objects:

```json
{
  "counter_id": 1,
  "object_ids": [1001, 1002],
  "from": 1620000000,
  "to": 1620086400,
  "aggregation": "avg",
  "group_by_objects": true
}
```

## Aggregation Methods

The @reportdb supports various aggregation methods:

- `avg`: Average of values
- `sum`: Sum of values
- `min`: Minimum value
- `max`: Maximum value
- `count`: Count of data points

## Caching

The @reportdb implements a caching system to improve query performance:

- Uses Ristretto cache for high-performance in-memory caching
- Caches frequently accessed data points
- Implements TTL (Time-To-Live) for cache entries
- Provides metrics for cache hit ratio monitoring

## Configuration

### Main Configuration

Configuration is stored in `config/config.json`:

```json
{
  "writers": 4,
  "readers": 4,
  "partitions": 8,
  "dataBuffer": 1000,
  "responseBuffer": 100,
  "eventsBuffer": 1000,
  "queryBuffer": 100,
  "dayWorkers": 10,
  "fileGrowthSize": 1048576,
  "saveIndexInterval": 300
}
```

- `writers`: Number of writer instances
- `readers`: Number of reader instances
- `partitions`: Number of partitions per day/counter
- `dataBuffer`: Size of the data channel buffer
- `responseBuffer`: Size of the response channel buffer
- `eventsBuffer`: Size of the events buffer
- `queryBuffer`: Size of the query channel buffer
- `dayWorkers`: Number of parallel workers per day
- `fileGrowthSize`: File growth size in bytes
- `saveIndexInterval`: Index save interval in seconds

### Counter Configuration

Counter definitions are stored in `config/counter.json`:

```json
{
  "1": {
    "name": "Memory Usage",
    "type": "uint64"
  },
  "2": {
    "name": "CPU Utilization",
    "type": "float64"
  },
  "3": {
    "name": "Hostname",
    "type": "string"
  }
}
```

- `name`: Human-readable name of the counter
- `type`: Data type (uint64, float64, string)

## Building and Running

### Prerequisites

- Go 1.18 or higher
- ZeroMQ 4.3 or higher
- Sufficient disk space for data storage

### Build Instructions

```bash
cd reportdb
go build
```

### Run Instructions

```bash
./reportdb
```

The Report Database will start and listen for data from the Backend System.

## ZMQ Communication

### Incoming Data (PULL Socket)

- **Socket**: tcp://*:6003
- **Format**: MessagePack-encoded array of Events objects
- **Content**: Metrics data with object ID, counter ID, timestamp, and value

### Incoming Queries (PULL Socket)

- **Socket**: tcp://*:6004
- **Format**: MessagePack-encoded QueryReceive objects
- **Content**: Query parameters including counter ID, object IDs, time range, and aggregation method

### Outgoing Results (PUSH Socket)

- **Socket**: tcp://*:6005
- **Format**: MessagePack-encoded Response objects
- **Content**: Query results with request ID and data

## Data Types

### Events

```go
type Events struct {
    ObjectId  uint32      `msgpack:"objectId" json:"objectId"`
    CounterId uint16      `msgpack:"counterId" json:"counterId"`
    Timestamp uint32      `msgpack:"timestamp" json:"timestamp"`
    Value     interface{} `msgpack:"value" json:"value"`
}
```

### Query

```go
type Query struct {
    CounterID      uint16    `msgpack:"counter_id" json:"counter_id"`
    ObjectIDs      []uint32  `msgpack:"object_ids" json:"object_ids"`
    From           uint32    `msgpack:"from" json:"from"`
    To             uint32    `msgpack:"to" json:"to"`
    Aggregation    string    `msgpack:"aggregation" json:"aggregation"`
    GroupByObjects bool      `msgpack:"group_by_objects" json:"group_by_objects"`
    Interval       int       `msgpack:"interval" json:"interval"`
}
```

### Response

```go
type Response struct {
    RequestID uint64      `msgpack:"request_id" json:"request_id"`
    Error     string      `msgpack:"error,omitempty" json:"error,omitempty"`
    Data      interface{} `msgpack:"data" json:"data"`
}
```

## Performance Characteristics

### Write Performance

- Optimized for high-throughput write operations
- Parallel processing with multiple writers
- Efficient file growth strategy
- Memory-mapped files for fast writes

### Read Performance

- Fast query execution with direct index lookups
- Parallel data fetching for improved throughput
- Caching for frequently accessed data
- Optimized binary data format

## Monitoring

The @reportdb provides monitoring capabilities:

- Cache hit ratio metrics
- Memory usage statistics
- Garbage collection metrics
- HTTP endpoint for pprof profiling (localhost:6060)

## Error Handling

The @reportdb implements comprehensive error handling:

1. **Communication Errors**: ZMQ socket errors are logged and operations continue when possible
2. **Storage Errors**: File and index errors are logged and reported
3. **Query Errors**: Invalid queries return error responses
4. **Graceful Shutdown**: Clean shutdown process ensures all resources are released

## Project Structure

```
reportdb/
├── config/
│   ├── config.json         # Main configuration
│   └── counter.json        # Counter definitions
├── src/
│   ├── bootstrap.go        # Application entry point
│   ├── cache/              # Caching implementation
│   ├── datastore/
│   │   ├── reader/         # Query processing
│   │   └── writer/         # Data writing
│   ├── logger/             # Logging configuration
│   ├── server/             # ZMQ server implementation
│   ├── storage/            # Storage engine
│   └── utils/              # Utility functions
└── README.md               # This documentation
```

## Future Improvements

- **Distributed Storage**: Support for multi-node operation
- **Compression**: Data compression for storage efficiency
- **Retention Policies**: Automatic data retention and cleanup
- **Query Optimization**: Advanced query planning and optimization
- **Replication**: Data replication for fault tolerance
- **Security**: Authentication and encryption
