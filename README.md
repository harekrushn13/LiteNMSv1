# LiteNMSv1

LiteNMS is a modular, scalable network management system designed for efficient device monitoring, data collection, and analytics. The project is composed of three main components:

- **Backend**: Central coordinator providing API endpoints, business logic, and communication with other system components.
- **Poller**: High-performance polling engine for collecting metrics from network devices.
- **ReportDB**: Custom time-series database optimized for storing and querying network metrics.

---

## Project Structure

- `backend/` - API server, business logic, and integration with poller and reportdb
- `poller/` - Device polling engine for metrics collection
- `reportdb/` - Time-series database for metrics storage and analytics

---

## Component Summaries

### Backend
- RESTful API built with Gin (Go)
- Handles device credential management, discovery, provisioning, and data queries
- Communicates with poller and reportdb via ZeroMQ
- Stores configuration and state in PostgreSQL
- Comprehensive error handling and logging

### Poller
- Collects metrics from provisioned network devices at scheduled intervals
- Implements prioritized scheduling and parallel execution
- Uses SSH for device communication
- Sends collected data to backend/reportdb via ZeroMQ
- Highly configurable and supports custom counters

### ReportDB
- Specialized time-series database for network metrics
- Efficient binary storage, partitioned by date and counter
- Supports gauge, histogram, and grid queries with aggregation
- Implements in-memory caching for fast queries
- Scalable, write-optimized, and supports parallel reads/writes

---

## Configuration

Each component has its own configuration files in their respective `config/` directories. See the module READMEs for details on environment variables and JSON config formats.

---

## Building and Running

See the `backend/README.md`, `poller/README.md`, and `reportdb/README.md` for detailed build and run instructions for each component.

---

## Documentation & Diagrams

- Each module contains a detailed README.
- Data Flow Diagrams (DFDs) and architecture diagrams are available in the project and module directories.

---

## Contribution

Contributions are welcome! Please see the module READMEs for guidelines and development setup.