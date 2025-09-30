# Installation of the indexer, setting up the database and running the indexer

## Installation of the indexer

You can install the indexer by building it from the source code or by using the docker image.

### Building from the source code

This will assume that you are using Debian based system. Just use dnf or zypper for other systems to install the
make package. It will also assume that you have the go installed.

```bash
sudo apt install make
git clone https://github.com/Cogwheel-Validator/spectra-gnoland-indexer.git
cd spectra-gnoland-indexer
make build
```

### Using the docker image

For now there aren't any prebuilt docker images. You can build your own by using the Dockerfile. It will also assume that you have the docker installed and that user has the docker group.
To build the image you can use the following command:

```bash
git clone https://github.com/Cogwheel-Validator/spectra-gnoland-indexer.git
cd spectra-gnoland-indexer
docker build -t cogwheel/spectra-gnoland-indexer:latest .
```

## Setup the database

The recommended way to do this would be to use docker for the database. You can also use any linux package to install the database. For this section it is better to head over to the official documentation of the database.

For linux https://docs.tigerdata.com/self-hosted/latest/install/installation-linux/.
For docker https://docs.tigerdata.com/self-hosted/latest/install/installation-docker/.

If you plan to use docker it would be better to use the timescaledb-ha image.

You can also use the docker-compose.yml file to setup the database. It is already configured to use the timescaledb-ha image.

The important thing to setup is the database connections. The indexer can do good job even with only 50 connections.
However if you are planning to add a lot of services a top of the database maybe increase it. The docker compose is 
set to 500 connections.

For creating all of the necessary tables and types you can use the setup.go file in the cmd directory.
The cmd has two commands to create the database and the user:

```bash
 collection of tools to set up and manage the database for the gnoland indexer.

Usage:
  setup [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  create-db   Create a new database named gnoland
  create-user Create a new user for the database
  help        Help about any command

Flags:
  -h, --help   help for setup
```

To use all of these commands you will need a user with admin privileges.
In the example below the user is postgres you will be asked for the password.
To create the database you can use the following command:

```bash
go run cmd/setup.go create-db --db-host localhost --db-port 5432 --db-user postgres --db-name postgres --ssl-mode disable --new-db-name gnoland --chain-name gnoland
```

To create the users for the database you can use the following command. 
The user is the name of the user to create and the privilege is the privilege level for the user.
The program will ask for admin password, and later it will ask for the password of the new user.
The privilege level can be "reader" or "writer". The reader should have only the select privileges.
The writer should have the select, insert, update and privileges.

```bash
go run cmd/setup.go create-user --db-host localhost --db-port 5432 --db-user postgres --db-name postgres --ssl-mode disable --user writer --privilege writer
```

## Running the indexer

The indexer can be ran in 2 modes: live and historic.

Now when you have the database running you can actually run the indexer. The indexer has a lot of flags that can be used to configure it:

```bash
A blockchain indexer for Gnoland that processes blocks and transactions.

Usage:
  indexer [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  historic    Run the indexer in historic mode
  live        Run the indexer in live mode

Flags:
  -e, --compress-events              compress events
  -c, --config string                config file path (default "config.yml")
  -h, --help                         help for indexer
  -m, --max-req-per-window int       max requests per window (default 10000000)
  -r, --rate-limit-window duration   rate limit window (default 1m0s)
  -t, --timeout duration             timeout (default 20s)
  -v, --version                      version for indexer
```

Compress events doesn't work at the moment so do not use it! 

The max request per window is rated to the max request that can be made to the RPC per rate limit window. So depending on the RPC rate limit option you can decrease the rate limit window and set up any desired value for the maximum requests per window.

The timeout is the timeout for the RPC requests. The default is 20 seconds. You can set it to lower if you want.

The config file is needed to configure the indexer. You can use the config.yml.example file as a template.
```yaml
# Example config file for the indexer

# RPC URL
# the indexer can listen to http or https
rpc: https://gnoland-testnet-rpc.cogwheel.zone

# Pool configuration
# these are settings related to the database connection pool
#
# This are some default values, you can change them to your needs
# maybe you can increase the min and max connections but this settings should be fine for most cases
# just be careful that postgres(timescale db) can accept those limits
pool_max_conns: 50
pool_min_conns: 10
pool_max_conn_lifetime: 5m
pool_max_conn_idle_time: 5m
pool_health_check_period: 1m
pool_max_conn_lifetime_jitter: 1m

# Indexer settings
# 
# Make sure the chain name is the same one you used when you created the database
# Chunk sizes are the maximum number of blocks or transactions that the indexer will process in a single chunk
# for now you can leave them as they are, you can increase or decrease them if you want
# 
# Reccomended chunk sizes are 50 blocks and 100 transactions but you should be safe to move block chunk size from 10 to 100
# and transaction chunk size from 10 to 200
chain_name: gnoland
max_block_chunk_size: 50
max_transaction_chunk_size: 100

# Live settings
# this is the time that the indexer will wait before it checks the chain if there are new blocks
live_pooling: 5s

# Retry settings
#
# These are settings related to the retry logic
#
# The retry amount is the number of times that the indexer will retry to get the blocks or transactions
# The pause is the number at which the indexer will make a bigger pause. 
# The indexer will use modulo operator on the retry attampts, if the retry attempt modulo pause is 0, the indexer will pause for the pause time
# The pause time is the time that the indexer will pause after failing to get the blocks or transactions
# The exponential backoff is the time that the indexer will wait before it retries to get the blocks or transactions
#
# The default values are 6 retries, 3 pauses, 15 seconds pause time, and 2 seconds exponential backoff
retry_amount: 6
pause: 3
pause_time: 15s
exponential_backoff: 2s
```

To run the indexer in historic mode you can use the following command:
```bash
indexer historic --config config.yml --from-height 1000 --to-height 2000
```
The from height is the starting height of the block to index. The to height is the ending height of the block to index.
The indexer will index the blocks from the from height to the to height inclusive. You can also add the other flags
such as the max request per window, the rate limit window, the timeout, etc.

Historic mode flags:
```bash
Runs the spectra indexer in historic mode, processing blocks from a given height to a given height.
The historic mode takes in starting height point and a finishing height. It should be used to 
sync up the database to the latest block height. 

It can also be useful if you want to index blockchain partially and work with data for any kind of testing
or partial scan of the chain where you want to index from a certain height to a certain height.

Usage:
  indexer historic [flags]

Flags:
  -f, --from-height uint   starting block height (default 1)
  -h, --help               help for historic
  -o, --to-height uint     ending block height (default 1000)

Global Flags:
  -e, --compress-events              compress events
  -c, --config string                config file path (default "config.yml")
  -m, --max-req-per-window int       max requests per window (default 10000000)
  -r, --rate-limit-window duration   rate limit window (default 1m0s)
  -t, --timeout duration             timeout (default 20s)
```

To run the indexer in live mode you can use the following command:
```bash
indexer live --config config.yml
```
Live mode flags:
```bash
Runs the spectra indexer, listening to any new blocks and processing them.
It will check the database for the last processed height and start from there
In the events the database is empty, it will start from block height 1. This can be used
to sync up the database to the latest block height
However if you do not need previous data, you can run the live mode with the skip-db-check flag set to true.
Afterwards you can run live mode normal without the skip-db-check flag.

Usage:
  indexer live [flags]

Flags:
  -h, --help            help for live
  -s, --skip-db-check   skip initial database check

Global Flags:
  -e, --compress-events              compress events
  -c, --config string                config file path (default "config.yml")
  -m, --max-req-per-window int       max requests per window (default 10000000)
  -r, --rate-limit-window duration   rate limit window (default 1m0s)
  -t, --timeout duration             timeout (default 20s)

```

You can also add the other flags such as the max request per window, the rate limit window, the timeout, etc.
The skip db check is a flag that will skip the initial database check. You can use it if you want to run the indexer from the latest chain height without previous data.

### When to use each mode and how to run it in the production

These mods can be used differently together. For example you might get access to the archive RPC node. But you 
want to run the indexer in segments. So you can use the historic mode to index the data in the database and then
when you reach some height that normal RPC node has access to you can switch to the live mode.

Maybe you need to view at some segment of the blockchain individually. You can set up different database within the 
same database server. From there you can instruct the indexer to index the data from a certain height to a certain 
height. And then you can look at the data from there and do any kind of testing or analytics for example.

If you started the indexer with the live mode without the prior data it will try to start the indexer from the 
height 1. If the RPC node doesn't have that height it will fail. So you can use the skip-db-check flag so that the 
live mode skips the check of what was the last processed height and will start from the latest chain height. This 
mode with skip db check is also useful for the testing of the indexer.

So the best way to have any kind of good setup would be to run the indexer in historic mode to index the data in 
segments. Then switch to the live mode to index the data in the real time. To be clear the live mode can process and
index the data the same as the historic mode, you just gain more control over the flow of the indexer.

### Deployment

