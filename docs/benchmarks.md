# Benchmarks

## Date 04/10/2025

The benchmarks were ran on different machines with different max block chunk size in the settings. It was done on
the old Gnoland 7.2 testnet from block height 950866 to 960866. The chunk size declares how much blocks can be
processed at the same time. The max transaction chunk was left at 100 as it is the default value.

| Cpu Count | 50 chunks | 100 chunks | 200 chunks | 500 chunks |
| 2 vCPUs AMD Epyc | 2m25s | 1m23s | 50s | 28s |
| 4 vCPUs AMD Epyc | 1m30s | 1m05s | 37s | 18s |
| 16 core/32 threads AMD Ryzen 9 7900x3d | 24s | 16s | 10s | 10-5.4s* |

* The 16 core/32 threads benchmark was ran 2/4 times the result was at 6.5s however at this point the RPC node might be a bottleneck. One time it got 10 seconds and another time it was 5.4 seconds.

This blocks are mostly empty of transactions so it might not be the best benchmark. But it is a good indication of the performance and can be used as a reference.

## Recommendations

For normal processing of the data, the recommended block chunk should be anywhere between 50-150. The transaction
chunk should be anywhere between 50-300. This is a good balance between the performance and the not causing the RPC
node to be overloaded.

If you are using your own RPC node, only dedicated for the indexer by default settings it should be able to handle
up to 900 requests concurrently. If you run the node with this settings you could in theory push the indexer to use
block chunk to size of 450 and transaction chunk to size of 900. However only do this if you are the only one using
the node because the node will simply block requests from other users, and even the indexer. The indexer will still
try to record the data and it will probably be successful, but it will actually be slower because it needs to retry
the missed requests. So unless you are the only one using the node, it is not recommended to use these settings.

Or if you are the owner of this node, you can adjust this parameter in the settings and allow for more requests.
If you plan to alter the node configurations better to consult the Gnoland documentation.
