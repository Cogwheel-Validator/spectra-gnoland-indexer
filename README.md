<h1 align="center"> The Spectra 💫 Gnoland Indexer</h1>
<div align="center">
<img src="./images/spectra-gnoland-indexer.svg" alt="Spectra Gnoland Indexer" width="250" style="padding: 25px;">
</div>

The Spectra Gnoland Indexer(SGI) is a tool that records the data from the Gnoland blockchain and stores the data 
in the timeseries database, Timescale DB. This indexer will be a part of the Spectra explorer and will be used to 
store the data for the explorer. This program can be used for any kind of data/analytics program or app.

The biggest problem when dealing with the blockchain data is how to store the data in a way that is easy to query 
and analyze. Some projects rely on the direct access to the data from the blockchain nodes and the problem in this 
case is that the nodes are meant for multiple purposes. They usually provide at least one, sometimes even more,
endpoints for querying the data, then this node need to be able to work with other nodes and to acquire the block 
data from other nodes via P2P network. Most nodes can be used as a miners/validators that need to sign the blocks, 
and maintain the blockchain state. While they are the first point of contact for data and contact with the chain
they are not most efficient when it comes to making some queries easy and fast. Plus the nodes are made to be fast 
and efficient when it comes to preformance of the nodes, so they are not most efficient when it comes to storing 
the data.

## Table of content
- [Table of content](#table-of-content)
- [The solution](#the-solution)
- [Why TimescaleDB? And can it work on other SQL databases?](#why-timescaledb-and-can-it-work-on-other-sql-databases)
- [The indexer and how it works](#the-indexer-and-how-it-works)
- [What is indexed and what is not](#what-is-indexed-and-what-is-not)
  - [Blocks:](#blocks)
  - [Validator signings:](#validator-signings)
  - [Transactions:](#transactions)
- [Pros and cons of the SGI](#pros-and-cons-of-the-sgi)
  - [🦾 Pros:](#-pros)
  - [🐞 Cons:](#-cons)
- [In depth documentation](#in-depth-documentation)

## The solution

The data in the blockchain is mostly tied to blocks, however when any kind of analytics is needed we need some 
parameter that falls familiar to us from the daily life. So it comes more natural to compare the data with some
sort of time parameter then to a block height. Then the data needs to be stored in a way that is easy to query and 
analyze. And since we are dealing with the time parameter, we need to have the ability to aggregate the data over 
some time period so we can timeseries and easily plot the data.

A lot of indexers use NoSQL databases for this use case. However they usually have their own problems and 
limitations. They are not as flexible as the SQL databases when it comes to the data types and the query language.
that is the reason why the TimescaleDB is a perfect fit for this use case.

Anyone with some experience with the SQL, especially with the PostgreSQL, can easily understand the TimescaleDB.
It is a extension of the PostgreSQL that adds the ability to store the time series data in a way that is easy to 
query and analyze.

## Why TimescaleDB? And can it work on other SQL databases?

TimescaleDB (Tiger Data is their commercial version) is a extension of the PostgreSQL that adds the ability to 
store the time series data in a way that is easy to query and analyze. It also has a lot of features that can 
extend the capabilities of the PostgreSQL and make it more powerful.

Technically speaking the indexer can work on Postgres database, however you would need to create the tables and 
types manually. This might be added later in the future if there is a demand for it but for now the focus is on the 
TimescaleDB.

There are some other SQL databases that could work in theory. For them to work they would need to have:
- The postgres wire protocol
- Equivalent of Numeric type ( some databases might have it as DECIMAL, but not all of them )
- The ability to create custom types

If all of the above are met then the indexer can work on them. It might work on CockroachDB for example but this is 
out of the scope of this project. Maybe in the future it might be an interesting idea to support it.

There were some other contenders for the database choice. They were: InfluxDB and QuestDB. 

At the time when the cosmos indexer was made the InfluxDB had a opensource version 2 but at the time it was lacking
some of the features that are now present in the version 3. Maybe it could have been used for the indexer but at 
the time it seemed to be not the best choice.

QuestDB was a interesting choice and it is probably even better choice than TimescaleDB. But it does have some 
limitations but nothing that would make it impossible to use it. However with the TimescaleDB it is pretty much an 
extention of Postgres and it is more mature and has more features and it can be extended with any other extensions 
that are available for the Postgres.

The TimescaleDB has also feature of it's own that is not present in the Postgres. The user could extend the indexer
by adding the data aggregation features, automatic jobs scheduling, hyperfunctions and more.

## The indexer and how it works

The indexer has 2 main modes of operation:
- Live mode
- Historic mode

Live mode is the mode that is used to index the data in the real time. It will sync up the database to the latest 
block height and will continue to index the data in the real time.

Historic mode is the mode that allows the user to index a certain range of blocks. For example you might not need 
the whole chain history or just need a part of it. This mode is useful for the testing, partial indexing of the 
chain or gradual indexing of the chain.

Durring both modes the indexer will use fan out method to send the requests to the RPC node. Then all of the data
is collected and processed and decoded in parallel. All of the addresses are collected and stored in the database 
and the indexer stores them in it's own cache. Then any address that is found in the transaction is referanced by 
the integer value in the transaction tables. At the end of the processing the data is inserted in batches into the 
database.

## What is indexed and what is not

The indexer stores the esential data that is related to the blocks, transactions, messages, and accounts.
It also stores the validator block signings and the validator addresses.
This way it is able to provide a lot of data for the analytics and the explorer.
The data is stored in the following tables:
- blocks
- transactions general data
- messages (each message type has its own table)
- regular and validator addresses (each address type has its own table)
- validator block signings
- validator addresses
- address transactions (to tie the addresses to the transactions)

Some of the data is not indexed and it is not planned to be indexed in the future. Such as:

### Blocks:
Stored data:
- Basic block hash
- Block height
- Block timestamp
- Block chain ID
- Block proposer address
- Block transactions hashes

Not stored data:
- Last commit hash
- App hash
- Data hash
- Validators hash
- Next validators hash
- Consensus hash

### Validator signings:
Stored data:
- Validator block signing height
- Validator block signing timestamp
- Validator block signing signed validators

Not stored data:
- Missed validators
- Precommits and all of the hashes and other data, so only a confirmation that the validator signed the block

### Transactions:
Almost all of the data regarding to the transaction and the messages are stored with the exception for the 
VM message Add Package and Call where in theory one could extract even the body of the smart contract. So this is unnecessary to store for explorers and other analytics tools. This might be added in the future if there is a demand for it.

## Pros and cons of the SGI

### 🦾 Pros:

- The indexer process the data using goroutines and channels, which can provide a faster processing of the data. Some early test on the real data showed about 2vCPU can index 10K blocks in about 2m30s while the 4vCPU can index 10K blocks in about 1m30s.
- The program has 2 modes that can be used for the indexing of the data. This can be useful for the testing, partial indexing of the chain or gradual indexing of the chain.
- The data is stored ready to be used for any kind of analytics and visualization with any programming language.
- No need to deal with Amino encoding and decoding for the messagess as the indexer decodes the messages and stores them in the database.
- It comes with a REST API to get you started quickly.(To be added in the future)
- It relies on a SQL database, which can provide a easier experience for any user that is familiar with the SQL.
- It uses TimescaleDB, a PostgresSQL extenstion that can be extended with any other extensions, plus the TimescaleDB has a lot of features that are not present in the Postgres.

### 🐞 Cons:

- The address cache provides fast reads but it intorduces a complexity. The developer that plans to use the indexer and plans to make some custom solution on working with the data will need to fully understand the data structure and how to use it. The REST API provies a easy way to interact with the data and to get the data in a readable format.
- The indexer relies on the RPC node for the data. If the RPC node is not available the indexer will not be able to index the data. ( although in the future the indexer might be able to use multiple RPC nodes )
- Technically the indexer has a limit of 2 billion addresses. If at any point the Gnoland grows to that size the indexer would need to be updated to support it. It is not a problem for now but it is something to keep in mind.

## In depth documentation

For more detailed documentation, please refer to the [docs](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/blob/main/docs/README.md) directory.