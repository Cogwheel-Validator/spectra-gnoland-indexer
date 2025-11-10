# Synthetic Integration Testing

This directory contains integration tests for the Spectra Gnoland Indexer using synthetic data.

## Overview

The integration tests use a **synthetic data generator** to create realistic blockchain data (blocks, transactions, validators) and test the entire indexer pipeline with a real database connection.

## Running Integration Tests

### Prerequisites

- PostgreSQL/TimescaleDB instance running
- Database schema initialized (run `go run indexer/main.go setup create-db`)
- Test configuration file created


Additional info, anything above 100K block height is not recommended 
to run. To clarify it can run but the
generator that creates the synthetic data stores the data in the RAM 
memory. Unless you have abnormally high RAM
you are risking to run out of memory. On avarage I saw the RAM going 
up to 4-5GB for the 100K blocks with 500 regular addresses and 50 
validators.

### Quick Start

1. **Create test configuration** (`test_config.yml`):

```yaml
host: localhost
port: 5432
user: postgres
password: 12345678 # do not use password default unless for development or testing!!!
dbname: gnoland
sslmode: disable
pool_max_conns: 50
pool_min_conns: 2
pool_max_conn_lifetime: 5m
pool_max_conn_idle_time: 2m
pool_health_check_period: 30s
pool_max_conn_lifetime_jitter: 30s

chain_id: gnoland
max_height: 1000
from_height: 1
to_height: 1000
```

2. **Run integration tests**:

```bash
# From project root
go test -v -tags=integration ./integration

# Or from integration directory
cd integration
go test -v -tags=integration

# With timeout
go test -v -tags=integration -timeout=10m ./integration
```

## What Gets Tested?

The integration tests exercise:

- **Orchestrator** - Block processing orchestration
- **Data Processor** - Transaction parsing and validation
- **Database Layer** - TimescaleDB insertion and querying
- **Address Cache** - Validator and address caching
- **RPC Client** - Response parsing (using synthetic data)

This provides **almost** entire **end-to-end validation** of the  indexer pipeline. The only thing that is not tested is the RPC client.
