# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0-alpha.2] - 2025-09-21

This is a second alpha release although the indexer is not yet ready. 

### Added

- Updated all of the dependencies to the latest version [5fb6b8d](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/5fb6b8dc07bcbacd5a8a66d4eb68a66435f2d695)
- Added the generator functions for the integration test [d61cfa6](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/d61cfa64088ad5654fa2553b7c77c56007451917), [34a46fa](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/34a46fafb40d762fc4ac256fd0605da15e6cba8b)
- Added the synthetic integration test [76e42f6](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/76e42f60b4a828a075322c35d03e8ab52a1721ea)
- Moved some of the code logic to it's own package [9cc12e9](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/9cc12e9961e5c7d2e984209faa5ffda97f75eb06), [76e42f6](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/76e42f60b4a828a075322c35d03e8ab52a1721ea), [9ca2214](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/9ca221475ac90df0edadc6b1eaf028feb75b79a6)


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
