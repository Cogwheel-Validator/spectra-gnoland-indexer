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
chain_id: gnoland-synthetic-1
max_height: 4000001
from_height: 1
to_height: 4000000
```

Max height is the maximum height of the synthetic chain. To height is the height upon which the test will stop.
Other parameters are self explanatory.