# REST API

With this indexer you can use the REST API from the api directory. It is built with chi framework and huma.

Out of the box you get a [Spotlight UI](https://stoplight.io/) to interact with the API on the /docs route. 
You can also use some other API UIs but you will need to make the changes yourself. See docs about changing the UI [here](https://huma.rocks/features/api-docs/).

The huma is framework agnostic and you could modify the API to use some other framework, use another middleware
or maybe use the stdlib http package. This API provides the most basic and necessary features for querying the database.

## API routes

There are total of 5 routes available. This are the basic routes that are needed for the most part. This might be expanded in the future.

### Blocks

- /block/{height} - Get a specific block data by height
- /blocks/{from_height}/{to_height} - Get a range of blocks data by height range
- /blocks/{block_height}/signers - Get all of the validators that signed that block + the proposer

### Transactions

- /transaction/{tx_hash} - Get a specific basic transaction data by hash, this gives the basic data about the transaction like hash, timestamp, block height, gas used, gas wanted, fee and more.
- /transaction/{tx_hash}/message - Get a specific transaction message data by hash, this gives more detailed data about type of transaction, specific data for that message type and more.

### Addresses

- /address/{address}/txs?from_timestamp={from_timestamp}&to_timestamp={to_timestamp} - Get all of the transactions for a given address for a certain time period

## Setup API

To setup the API you can use the config file. The example config file is in the root under config-api.yml.example.
```yaml
# Example config file for the API
host: 127.0.0.1
port: 8080
cors_allowed_origins:
  - "*"
cors_allowed_methods:
  - "GET"
cors_allowed_headers:
  - "Origin"
  - "Content-Type"
  - "Accept"
cors_max_age: 600
chain_name: gnoland
```

Some of the environment variables are located under the .env file. The example .env file is in the root under .env.example.

```env
# Example .env file for the API
# do not use password default unless for development or testing!!!
API_DB_HOST=127.0.0.1
API_DB_PORT=5432
API_DB_USER=reader
API_DB_SSLMODE=disable
API_DB_PASSWORD=12345678
API_DB_NAME=gnoland

# these are the default values for the database connection pool
# if they are not filled the API will load the default values
API_DB_POOL_MAX_CONNS=50
API_DB_POOL_MIN_CONNS=10
API_DB_POOL_MAX_CONN_LIFETIME=10s
API_DB_POOL_MAX_CONN_IDLE_TIME=5m
API_DB_POOL_HEALTH_CHECK_PERIOD=1m
API_DB_POOL_MAX_CONN_LIFETIME_JITTER=1m
```

You can make the API by running the following command from the project root:

```bash
make build-api
```
This command will build the API and it will be located in the build directory.

To run the API you can use the following command:

```bash
./build/api -c config-api.yml
```

You can also use the following command to run the API with HTTPS if you have the cert and key files:

```bash
./build/api -c config-api.yml -t cert.pem -k key.pem
```

The docker image doesn't exist for now and it will be added in the future.

