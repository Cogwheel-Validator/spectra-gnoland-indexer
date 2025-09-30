# Requirements

The indexer could can run on one machine while the database is located on another machine. You can also use the 
Tiger Data (Timescaledb cloud edition) which is a managed service by Tiger Data which is a company that builds the 
Timescaledb.

However for the ease of understanding through the documentation will refer that the indexer and the database are located on the same machine.

## Hardware requirements

To run the indexer you need to have the following system requirements:

Minimum system requirements:

- 2vCPUs
- 8GB RAM

Recommended system requirements:

- 4vCPUs
- 16GB RAM

The indexer could probably run on ARM64 architecture but it is not tested yet. So stick with the x86_64 architecture.

For the storage it can work on the HDD, but if the Gnoland chain grows in popularity and number of users increases, 
the HDD might not be a good choice in the long term. So maybe a regular SATA SSD could be a better choice. 
But if you are just running some testing or just a small deployment, the HDD will do the job most of the time. 
As for the size it depends on the amount of the data that you are indexing. 
For example the integration test did about 400K blocks and 1 million 
transactions. So it took around 2.2 GB of disk space. There are a lot of things that can affect the size of the 
database but at least use this as some reference point to how much space you might need.

For the RAM and CPU it kinda depends but for now this is a good starting point. As the database size grows, you 
might need to increase the RAM and CPU.

## Software and OS requirements

The following software and OS requirements are required to run the indexer:

- Go 1.25.1
- TimescaleDB 2.18 or higher but with PostgresSQL 16 or higher
- OS: Linux, anything based on Debian(Ubuntu, Mint, etc.) or RHEL(CentOS stream, Rocky Linux, etc.) should work, openSUSE also ok
- Docker (optional)