# Synthetic Integration Testing

This test uses a synthetic query operator and with generator package creates synthetic data for the indexer to process. 

## Requrements to run the test

- A TimescaleDB already setup and running with the same schema as production
- Go to run the main.go to run the test
- test_config.yml is the configuration file for the test (check the test_config.example.yml for the configuration)

To have a ready database for the test, run the setup.go file in the cmd directory. In the events you want to
integrate the indexer into your existing TimescaleDB production database, for now you would need to create the
tables and types manually untill the setup.go file is updated to create the tables and types automatically.

The config file is used to configure the database connection and set up the test parameters.

Additional info, anything above 100K block height is not recommended to run. To clarify it can run but the
generator that creates the synthetic data stores the data in the RAM memory. Unless you have abnormally high RAM
you are risking to run out of memory. On avarage I saw the RAM going up to 4-5GB for the 100K blocks with 500 regular addresses and 50 validators.

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

# synthetic test config
chain_id: gnoland
max_height: 100000
from_height: 1
to_height: 100000
```

Max height is the maximum height of the synthetic chain. To height is the height upon which the test will stop.
Other parameters are self explanatory.

## Historic test

Historic test is a test that runs the indexer in historic mode. It will generate the synthetic data and process it in the database. The test generates fictive data starting from the september 1st 2025, midnight UTC. The block production rate is 1 block per 5 seconds. It will generate 500 regular addresses and 50 validators so it doesn't need to generate a lot of new addresses. Depending on the specs of your machine the test may take a while to complete. For example 100K blocks test may take anywhere from the 30 minutes to 1 hour to complete. Moslty the program will spend time generating the synthetic data. The part where it insert the data into the database is at most a couple of minutes.

To run the historic test, run the historic.go file in the integration directory.

```bash
go run historic.go
```
