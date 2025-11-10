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

- The indexer doesn't insert the account txs into the database. (Solved)
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

## Date from 2025-09-28 to 2025-09-29

### Test configuration

This test was ran a bit differently. So it wouldn't crash due to storing it all in the RAM memory I ran it in chaunks of 100 K blocks 4 times making it 400 K blocks. So technically there are 200 different validators and around 2000 regular addresses. While probably not the most accurate test it was a good way to test the integration test with a lot of data.

- **Chain ID**: gnoland
- **Max height**: 400000
- **From height**: 1
- **To height**: 400000
- **Block production rate**: 1 block per 5 seconds
- **Regular addresses**: 500
- **Validators**: 20
- **Number of transactions**: 1,056,226

### Process information

It took way longer that it should have due to the bug in the synthetic query operator It was trying to generate on every section from block height 1. This was fixed later, leading back to the normal time.

- **Data generation duration**: 8+ hours but due to the bug
- **Data insertion duration**: 2-5 minutes per each run
- **Database size**: total for all runs around 2.3 GB

### Bugs

- The synthetic query operator was trying to generate on every section from block height 1. This was fixed later, leading back to the normal time. (Solved)
- The address tx table recevies duplicate data. Meaning that the map that pulls the addresses from the transaction was not working as expected. (Solved)
- The generator created only one validator address. (Solved)

### Known limitations

- The program doesn't have support for compressing the events and arguments. It only supports native format.
- When the test runs any generated data is stored in RAM memory. The more you have blocks and addresses the more RAM you need. With the current setup 100K blocks and 500 addresses it takes around 4-5GB of RAM.

### Summary

The test was mostly a success. The new bug was found. I will fix it later but it probably happened because I tried to store the data in a map and treat it like a set. I will need to look into this more.

The database holds around 2.3 GB of data, out of which:
| transaction_general     | 584 MB                |
| address_tx              | 357 MB                |
| bank_msg_send           | 139 MB                | 
| validator_block_signing | 139 MB                |
| vm_msg_run              | 124 MB                |
| blocks                  | 113 MB                |
| vm_msg_call             | 109 MB                | 
| vm_msg_add_package      | 50 MB                 |
| gno_addresses           | 216 KB                |
| gno_validators          | 16 KB                 |

The transaction general stores the most, I think this is due to events. Althoug this was synthetic representation of the data it still stores a lot of data. Untill I run the program with real data I will not know the real picture of how much do events pop up into the transactions and what kind of real data they store. This could be decreased with compressing the events but I would need to train the zstandard dirctionary for this to be efficient.

Address tx stores the second most, but I think this was due to the bug. On average every transaction had at least one more doubled entry from the signers. I will need to fix it later. But out of 2.5 million entries there is at least 1 million duplicated entries. So in reality this should be at least 40% less. I do not think there is much improvement for space here.

Bank msg send stores the third most, this is expected as it stores the data from the bank msg send. But this is moslty because the chances for the generator to create this type of transaction is higher than the other types.

Validator block signing stores the fourth most, this is expected as it stores the data from the validator block signing. However it only stored one validator address and it all pointed to that validator address. Regardless of the bug, in realaty this could get a bit bigger but not much.

Most of the vm messages I guess were ok but without looking at the real data and running the indexer I won't know for sure. The only improvment here would be for arguments in the vm_msg_call to be compressed for some longer arguments.

Regular addresses were stored but the validator addresses were not. I will need to fix this later.

Overall the test was a success. There were bugs related to the synthetic data generator and the indexer.