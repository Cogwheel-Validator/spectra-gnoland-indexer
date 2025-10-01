# Scailing the indexer

If you use the Tiger Data (TimescaleDB cloud edition) they have some internal methods of scaling the database that 
are not present for the self hosted version.
Regardless the Timescaledb is still a Postgres database and there are some ways to scale the indexer.
I will just mention some of the methods that are possible, I will not go into the details of how to do it. I might make some notes in the future for this section.

The Postgres database can scale vertically and horizontally(although the horizontal scaling is more complex).
With the vertical scaling you would need to increase the resources of the database server. With the horizontal 
scaling you would need to increase the number of database servers. The horizontal scaling is more complex because
you would need to handle the way that data is extracted and inserted into the database.

## Read replicas

The Postgres has a feature of read replicas. This can be used to scale the indexer if there are a lot of read 
operations. The read replicas are a copy of the database that is updated asynchronously. You would need to set up
all of the read replicas to gather the data from the master node. This is a bit out of scope for this project.

## Sharding

Sharding is also possible but it is a bit more complex for the indexer. There are a lot of methods to do this, I 
would recommend to do application level sharding. The proxy sharding is also a valid option. The 
catch with the indexer is that you would need to have some stop point at which you would split the data. The best 
would be either by time or by block height. The indexer doesn't have a stop option for the live indexing so you 
would need to have some other way to stop the indexing. This might be changed in the future but for now there is no 
such option. The indexer historic mode has a stop option by height but this can only be done for the blocks that 
have already been produced by the blockchain.

The thing that you would need to pay most attention is to the address cache. The indexer has in memory cache that
ties the address to the integer value and are mapped everywhere where some sort of adderess is stored.
So for this to work you might need to copy all of the addresses from one database to another. I guess you could 
also skip this part but then this would require some sophisticated way to query from all of the shards.