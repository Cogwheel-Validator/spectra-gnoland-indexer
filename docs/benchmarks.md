# Benchmarks

## Date 04/10/2025

The benchmarks were ran on different machines with different max block chunk size in the settings. It was done on
the old Gnoland 7.2 testnet from block height 950866 to 960866. The chunk size declares how much blocks can be 
processed at the same time. The max transaction chunk was left at 100 as it is the default value.

| Cpu Count | 50 chunks | 100 chunks | 200 chunks | 500 chunks |
| 2 vCPUs AMD Epyc | 2m25s | 1m23s | 50s | 28s |
| 4 vCPUs AMD Epyc | 1m30s | 1m05s | 37s | 18s |
| 16 core/32 threads AMD Ryzen 9 7900x3d | 24s | 16s | 10s | 10-5.4s* |

* The 16 core/32 threads benchmark was ran 2/4 times the resault was at 6.5s however at this point the RPC node might be a bottleneck. One time it got 10 seconds and another time it was 5.4 seconds.

This blocks are mostly empty of transactions so it might not be the best benchmark. But it is a good indication of the performance and can be used as a reference.

However for the safety of the indexer I would recommend to use at the 50 - 100 chunks for now. Having 200 is a bit risky at this point untill more testing is done.





