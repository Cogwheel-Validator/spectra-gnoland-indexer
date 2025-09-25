# Historic Report

## Date from 2025-09-21 to 2025-09-25

The test was ran multiple times so the data is not exactly the same. Some first tries failed until some major bugs were fixed.
Mostly it is a success but there are still some bugs that need to be fixed.

### Test configuration

- **Chain ID**: gnoland
- **Max height**: 100000
- **From height**: 1
- **To height**: 100000
- **Block production rate**: 1 block per 5 seconds
- **Regular addresses**: 500
- **Validators**: 50

### Process information

Per each run it took:

- **Data generation duration**: 45-60 minutes
- **Data insertion duration**: 2-5 minutes
- **Database size**: 670 - 840 MB

### Bugs

- The indexer doesn't insert the account txs into the database. (Not solved yet, HIGH)
- The indexer had problem to insert data properly because the database package had mixed up the data when it needed to pull the data from the database. (Solved)
- Pgx needed to have a some data types in pxg.Array type. (Solved)
- Pgx needs to have any kind of new type registered. Otherwise it will throw an error. (Solved)
- When the program started with the "historic" process it would add +1 to the chunk end height. The reasult would cause to have interlaping chunks and the program would throw an duplication error and would rewert any of the data that was inserted for that chunk. (Solved)
- Numeric type needs to use pgtype.Numeric type. Otherwise it will throw an error. (Solved)

### Known limitations

- The program doesn't have support for compressing the events and arguments. It only supports native format.
- When the test runs any generated data is stored in RAM memory. The more you have blocks and addresses the more RAM you need. With the current setup 100K blocks and 500 addresses it takes around 4-5GB of RAM.
- The data recorded was barely for one week. So TimescaleDB didn't use any kind of compression.

### Summary

Size of the database may varry depending but for this test it was around 670 - 840 MB. The test was somewhat successful, most of the bugs have been fixed but the indexer doesn't insert the account txs into the database.
Still some data is missing, but the data that is present is correct and it preformed within reasonable times.
When the fix for the account txs is implemented I will try to run the integration test for larger period so at least 1 month of data is recorded.