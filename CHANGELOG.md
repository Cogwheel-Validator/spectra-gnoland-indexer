# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0-alpha.1] - 2025-09-15

This is the first alpha release of the Spectra Gnoland indexer. Technically most of the indexer components are
done but it is not tested fully so this version is not recommended for production use. 

### Added

- CLI for the Indexer
- Config and env loaders
- RPC client with rate limiting
- PGX pool Postgres Client
- Address cache for regular and validator addresses
- Signal hook for graceful shutdown and emergency shutdown
- Amino decoder for the data from the Gnoland Chain
- Major operator/worker pattern for the indexer have been implemented
- Basic database setup 

### Known Issues

- The indexer is not tested and it is not recommended for production use.
- The setup program only sets the database and ties it to the admin user. This could be bad for security.
- The proto encoding for the events is not tested yet and might not even end in the final release.
- Zstandard compression has been added but it has only been used in some minor test nothing more. For this to work properly a synthetic dataset would need to be created and used to train the dictionary. Alternatively it can be trained on the real data but given that the chain is still in the development stage there is no gurantee it will have enough data to train a good dictionary.
