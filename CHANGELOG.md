# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0-beta.2] - 2025-09-29

The indexer had some bug fixes and some small improvments. The integration test was technically successful but there seems there is some kind of bug with the indexer. The indexer is not fully tested yet only the historic process has been tested. But not any runs were made on the real data. You can try to run this version on the real data but be advised it is not fully tested and might not work as expected.

### Added

- Makefile has been added. If you feel advanterous you can try to build the indexer with greentea garbage collection. [9fdad03](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/9fdad03ca3b9fe213dced5e1ef68912cc792355a)
- Apperently the previous versions didn't had the method to insert the data for the table address_tx. Now every transaction that was executed can be tied to each address that was involved in the transaction. [9fdad03](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/9fdad03ca3b9fe213dced5e1ef68912cc792355a), [6dd764](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/6dd76464b68809ac8df63f2c66f11678e1083b14)
- The CLI for the database setup now has a new command to create a new user for the database and appoint privileges to the user. It can be a reader(for APIs and some other programs that need SELECT privileges) or a writer(example indexer for historical data). [c39c1f7](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/c39c1f7b5f992da468710a54401f73efa6611881)
- Added a retry mechanism for the query operator. [900ee4f](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/tree/900ee4ff933e1015acc7f9a80de28201075370cf)

### Changed

- Updated the go version to 1.25.1 [b3e02b0](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/b3e02b0dc4f7f1896480c6e2c80ccecf79bbb1be)

## [0.1.0-beta.1] - 2025-09-25

The indexer had some bug fixes and some small improvments. The integration test was technically successful but there seems there is some kind of bug with the indexer. The indexer is not ready for production use.

### Changed

- When the indexer decodes the data using Amino decoder it unloads the data into a map[string]any, then from there it would make 2 conversions, one for the general data struct and the second for the sql data types. The idea was to have seperated logic for the general data struct and sql types. However at this point the indexer already needs to call the copy from method where the data is again being unloaded into some sort of tuple. So the first conversion was removed. [50ca1f2](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/50ca1f2e0d3ee1a3637ca26cdd70e5b48732da8d)
- Updated all of the dependencies to the latest version [5370a5c](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/5370a5c5486be5ef3803f16f968c383598e7f033)

### Fixed

- Fixed the sql related bugs, added some missing types, switched to pgtype.Numeric for the amount type [8aea191](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/tree/8aea1919ad7c3ad16c75a4bd2d1afe934a810dc4), [2b7ed52](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/2b7ed528e23c52c2849d2731cd187e921bf6223e),[ddfdcc1](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/ddfdcc1955784ad510de7f7c847d1a8cf3009e71)
- In some instances the pgx data need to be in the pgtype.Array for instance Txs for the block need to be stored into the pgtype.Array. The indexer now uses a generics function to convert the data into the pgtype.Array [2b7ed52](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/2b7ed528e23c52c2849d2731cd187e921bf6223e)
- The chunk end height was incremented by 1 when the indexer started the historic process. This caused the chunks to overlap and the indexer to throw an error about the duplication. The indexer now correctly sets the chunk end height to the max height [ddfdcc1](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/ddfdcc1955784ad510de7f7c847d1a8cf3009e71)
- Fixed a bug where the data processor would ask the address from the regular address cache instead from the validator address cache [ddfdcc1](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/ddfdcc1955784ad510de7f7c847d1a8cf3009e71)


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
