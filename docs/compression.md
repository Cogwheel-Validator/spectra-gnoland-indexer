# Event Compression

WARNING: This feature is still experimental and it is not recommended for production use. It might even be removed
in the future if there is no use for it.

The Spectra Gnoland Indexer does allow to compress the transaction events to reduce the size. This is done by using
the zstandard compression algorithm combined with protobuf serialization.

## Why this combination?

This is a mechanism that is used for the Spectra indexer for Cosmos SDK chains. The idea is to reduce the total size
as much as possible. The protobuf can serialize the data into very compact format and it uses logic that "has
repetition". Every part of the event has it's own slot in protobuf that is usually marked using numbers, which are
later serialized into bytes. This allows for additional compression by the zstandard algorithm. The zstandard
can train the dictionary to allow further compression on small data. The events are usually not very long but some
transactions can harbor a lot of events and each event can have a lot of attributes. This can add up to a lot of
data. But if the data from the event is serialized with protobuf and then compressed with zstandard, the final
size can be reduced by a lot.

## For what use cases this might work?

Depending on what you plan to do with this data this might not work for your use case. If the purpose is simply to
present the data as it is then you should be fine such as blockchain explorer, visual dashboards, etc.
If you plan to do heavy analytics on the data such as aggregating data, this can be a big
problem. The compressed data can't be queried over SQL in the database so any data you need you will need to
decompress and deserialize the data into the native format. This can be a big performance hit if you are doing
heavy analytics on the compressed data.

## How to use this feature?

At the current stage the trained zstandard dictionary exists. It should be possible to add `-e or --compress-events`
flag when running the indexer in live or historic mode. This will tell the indexer to compress the events.
However this feature is not fully ready.

The current dictionary is not production ready. It was trained from the real data but the sample was small. To
gain any significant compression the sample needs to be much larger.

## How to train the dictionary?

WARNING this process produces totally new zstd dictionary. You must keep this file safe and not lose it. Any
compression done with this dictionary will only work with this dictionary. IF YOU LOSE IT, YOU WON'T BE ABLE TO
TO RECREATE THE DICTIONARY AND THE COMPRESSED DATA WON'T BE ABLE TO BE DECOMPRESSED! Always try to use the official
dictionary unless you really know what you are doing. Always commit this dictionary to the repository, that way you
can always recreate the dictionary if you need to. More on zstd dictionary can be read at this [repo](https://github.com/facebook/zstd#the-case-for-small-data-compression).

To simplify the process run the indexer first and collect as much data as possible. The training program will
collect the data from the database and then build the zstd dictionary from the events.

In the project root directory you will find the `training-config.example.yml` file. This file is used to configure
the training process. Copy it to `training-config.yml` and configure it to your needs.

The dictionary can be trained by using the `make train-zstd` command.

There is a hard cap on amount of transaction that can be used to train the dictionary is set  to 250000. The
dictionary will be placed in `pkgs/dict_loader/events.zstd.bin` file.
