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
data. But if the data from the event is serialized with protobuf

## Why this might not work on Gnoland?

Gnoland is unique, it does share some similarity with Cosmos SDK chains but most of the logic lives in the VM.
For now there hasn't been any major testing done for this feature. Testing this in a synthetic database could
be a bad way to test this feature because the synthetic data is not representative of the real data. The real
data is much more complex and it is not easy to create a synthetic dataset that is representative of the real data.
For now the focus is to get the feature working and then test it on the real data.

Also for this to work and to have any real effect, the final serialized protobuf data would need to hold at least 80 bytes, in it's serialized form, to be worth the compression. It is not a hard limit but it is a good rule of
thumb. With trained dictionary this could possibly be even lower.
Some Gnoland transactions do have over 10 events that are complex, and then there those with 1 or none at all.

## For what use cases this might work?

Depending on what you plan to do with this data this might not work for your use case. If the purpose is simply to
present the data as it is then you should be fine. If you plan to do heavy analytics on the data this can be a big
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

The dictionary can be trained by using the `compression/cmd/main.go` file.

```bash
go run compression/cmd/main.go --config training-confg.yml --amount 10000 --chain-name gnoland --dict-path events.zstd.bin
```

There is a hard cap on amount to be collected to 50000. Due to previous experience with this feature, the data
collected anywhere from 50K to 250K events is enough. Theoretically you can expect anywhere between 50K events up to
200K events to train on. You could change the limit in the code if you want just be prepared to have enough RAM.
Anything over 1 million do not expect any significant compression improvement.
