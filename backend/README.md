# Backend Component - AugmentLNMS

The Backend component serves as the central coordinator for the AugmentLNMS system, providing API endpoints, business logic, and communication with other system components.

## Architecture Overview

The Backend follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                      API Layer (Gin)                        │
└───────────────────────────────┬─────────────────────────────┘
                                │
┌───────────────────────────────▼─────────────────────────────┐
│                    Controllers Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Credential │  │  Discovery  │  │      Provision      │  │
│  │ Controller  │  │ Controller  │  │     Controller      │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│                                                             │
│  ┌─────────────────────────┐                                │
│  │    Query Controller     │                                │
│  └─────────────────────────┘                                │
└───────────────────────────────┬─────────────────────────────┘
                                │
┌───────────────────────────────▼─────────────────────────────┐
│                     Data Layer (SQL)                        │
└───────────────────────────────┬─────────────────────────────┘
                                │
┌───────────────────────────────▼─────────────────────────────┐
│                   Communication Layer                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Polling   │  │     DB      │  │       Query         │  │
│  │   Server    │  │    Server   │  │       Server        │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Key Components

### API Layer

- **Gin Router**: HTTP router based on the Gin framework
- **API Routes**: Defined routes for different API endpoints

### Controllers Layer

- **Credential Controller**: Manages credential profiles for network devices
- **Discovery Controller**: Handles device discovery using SSH
- **Provision Controller**: Manages device provisioning for monitoring
- **Query Controller**: Processes data queries and returns results

### Data Layer

- **SQL Database**: PostgreSQL database for storing configuration and state
- **Models**: Data structures representing database entities

### Communication Layer

- **Polling Server**: Communicates with the Polling Engine via ZMQ
- **DB Server**: Sends metrics data to the Report Database via ZMQ
- **Query Server**: Exchanges queries and results with the Report Database via ZMQ

## Communication Channels

- **deviceChannel**: For sending provisioned device information to the Polling Engine
- **dataChannel**: For receiving metrics data from the Polling Engine and forwarding to the Report Database
- **queryChannel**: For sending query requests to the Report Database
- **queryMapping**: Maps query IDs to response channels for handling asynchronous responses

## API Endpoints

The Backend provides the following RESTful API endpoints:

### Credential Management

- `POST /lnms/credentials/`: Create a new credential profile

```json
{
  "username": "admin",
  "password": "password123",
  "port": 22
}
```

### Device Discovery

- `POST /lnms/discoveries/`: Create a new discovery profile

```json
{
  "credential_ids": [1, 2],
  "ip": "192.168.1.1",
  "ip_range": "192.168.1.0/24"
}
```

- `GET /lnms/discoveries/:id`: Start a discovery process

### Device Provisioning

- `POST /lnms/provisions/`: Provision a device for monitoring

```json
{
  "ip": "192.168.1.1",
  "credential_id": 1,
  "discovery_id": 1
}
```

### Data Querying

- `POST /lnms/query/`: Query metrics data

```json
{
  "object_id": 123,
  "counter_id": 456,
  "start_time": 1609459200,
  "end_time": 1609545600
}
```

## Data Flow

### Discovery Flow

1. Client sends discovery request to API
2. Discovery Controller validates request
3. Discovery Service creates profile in database
4. Discovery Runner establishes SSH connections to network devices
5. Discovery results stored in database
6. Results returned to client

### Provisioning Flow

1. Client sends provisioning request to API
2. Provision Controller validates request
3. Device information stored in database
4. Provisioned device sent through deviceChannel to Polling Server
5. Polling Server forwards device information to Polling Engine via ZMQ
6. Confirmation returned to client

### Query Flow

1. Client sends query request to API
2. Query Controller validates request
3. Query sent through queryChannel to Query Server
4. Query Server forwards query to Report Database via ZMQ
5. Results returned via ZMQ to Query Server
6. Query Server routes results to appropriate response channel
7. Query Controller returns results to client

## Configuration

### Configuration File

Configuration is stored in `config/config.json`:

```json
{
  "deviceBuffer": 5, // Buffer size for device channel
  "dataBuffer": 5, // Buffer size for data channel
  "queryBuffer": 50, // Buffer size for query channel
  "dbHost": "localhost",
  "dbPort": "5432",
  "dbUser": "postgres",
  "dbPassword": "postgres",
  "dbName": "LiteNMS",
  "dbSSLMode": "disable"
}
```

## Building and Running

### Prerequisites

- Go 1.18 or higher
- PostgreSQL 13 or higher
- ZeroMQ 4.3 or higher

### Build Instructions

```bash
cd backend
go build
```

### Run Instructions

```bash
./backend
```

The server will start on port 8080 by default.

## Dependencies

- [Gin](https://github.com/gin-gonic/gin): HTTP web framework
- [sqlx](https://github.com/jmoiron/sqlx): Extensions to database/sql
- [ZeroMQ](https://github.com/pebbe/zmq4): Messaging library
- [zap](https://github.com/uber-go/zap): Structured logging
- [msgpack](https://github.com/vmihailenco/msgpack): MessagePack encoding

## Project Structure

```
backend/
├── config/
│   └── config.json         # Configuration file
├── controllers/            # Business logic controllers
│   ├── credential.go       # Credential management
│   ├── discovery.go        # Device discovery
│   ├── provision.go        # Device provisioning
│   └── query.go            # Data querying
├── logger/                 # Logging utilities
├── models/                 # Data models
│   ├── credential.go       # Credential model
│   ├── db.go               # Database initialization
│   ├── discovery.go        # Discovery model
│   └── provision.go        # Provision model
├── routes/                 # API route definitions
│   └── routes.go           # Route initialization
├── server/                 # ZMQ server implementations
│   ├── polling.go          # Polling server
│   ├── reportdb.go         # DB server
│   └── query.go            # Query server
├── utils/                  # Utility functions
│   ├── config.go           # Configuration loading
│   ├── db.go               # Database utilities
│   └── types.go            # Common type definitions
├── main.go                 # Application entry point
└── README.md               # This documentation
```

## Error Handling

The Backend implements comprehensive error handling:

1. **API Level**: Returns appropriate HTTP status codes and error messages
2. **Controller Level**: Validates input and handles business logic errors
3. **Database Level**: Handles database connection and query errors
4. **ZMQ Level**: Handles communication errors with external systems
5. **Graceful Shutdown**: Properly closes connections and channels on shutdown

## Logging

The Backend uses the zap logging library for structured logging:

- **Info**: Normal operation information
- **Warn**: Potential issues that don't affect operation
- **Error**: Errors that affect operation but don't cause shutdown
- **Fatal**: Critical errors that cause shutdown

## Security Considerations

- Credentials are stored in the database
- API endpoints should be secured with authentication in production
- Database connection should use SSL in production
- Environment variables should be used for sensitive configuration

## Future Improvements

- Add authentication and authorization
- Implement rate limiting
- Add metrics and monitoring
- Implement caching for frequently accessed data
- Add support for HTTPS
