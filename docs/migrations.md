# Migrations

Since v1.0.0 the database schema is versioned. Every schema change ships as a numbered migration inside the
indexer binary, and the versions applied to a database are recorded in the `schema_migrations` table.

| Version | Release | Change                                                                                                                                                                                                                                                                                          |
| ------- | ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1       | v0.10.1 | Baseline, the schema created by v0.10.1. Nothing to run.                                                                                                                                                                                                                                        |
| 2       | v1.0.0  | `gno_addresses` / `gno_validators` keep one row per address for all chains (`chain_name` replaced by a `chains` array, unique key on `address`), new `chain_stats` table, `resolve_addresses` / `resolve_validators` functions, `tx_counter` gains `failed_count`, `gas_used` and `gas_wanted`. |

## How it works

- `setup create-db` creates the latest schema and records every version as applied.
- `run live` / `run historic` read `schema_migrations` on start and stop when a version is missing, or when the
  database has a version the binary doesn't know (a newer indexer migrated it). Pass `--skip-migration-check`
  to only log a warning instead, for forks that manage their own schema.
- `migrate up` applies the missing migrations in order. Each runs in its own transaction together with its
  `schema_migrations` row, so a failing migration is rolled back completely and recorded with `success = false`.
  An advisory lock keeps two `migrate up` runs from overlapping.

## Upgrading a v0.10.1 database to v1.0.0

Databases created by v0.10.1 have an empty `schema_migrations` table. `migrate up` recognizes the v0.10.1
schema, records version 1 as the baseline and then applies version 2.

1. Stop the indexer (and the API, if it writes API keys).
2. Back up the database.
3. Check the state:

    ```bash
    indexer migrate status --db-host localhost --db-port 5432 --db-user postgres --db-name gnoland
    ```

4. Migrate, with the account that owns the tables (the one used for `setup create-db`):

    ```bash
    indexer migrate up --db-host localhost --db-port 5432 --db-user postgres --db-name gnoland
    ```

5. Start the v1.0.0 indexer.

Like the other commands, the password is prompted for or read from `DB_PASSWORD`.

Notes on migration 2:

- Existing address and validator ids are kept, so the hypertables don't change.
- In v0.10.1 the same address indexed on two chains got two ids. Merging them would mean rewriting every row
  that references them, so the migration refuses such a database. In that case create a new database with
  `setup create-db` and re-index.
- `tx_counter` is dropped and recreated with the same policies, then refreshed over the full history after the
  commit. On a large database the refresh takes a while; if it fails the schema is still migrated, real-time
  aggregation keeps the results correct, and you can rerun it with `setup refresh-aggregates`.
- Privileges other roles hold on `gno_addresses` are copied to `chain_stats`, and the grants on `tx_counter`
  are restored, so existing `reader` / `writer` users keep working.
- If the database held multiple chains, and if there are duplicated addresses ( each tied to chain_name of the
  chain), then this migration won't be possible to be executed. The reason is each new address would need to be
  merged and then every insert would need to be checked and altered. Depending on the amount of data inserted
  it can take a lot of time to check and adjust the data. It is easier and safer to reindex in that case. This
  should be the last breaking change in terms of schema changes that won't have a proper plan to migrate.

## Databases without a recorded version

`migrate up` only recognizes the v0.10.1 schema automatically. Databases created before v0.10.0 can't be
migrated. For a database that is known to have a later schema but was never versioned (for example one created
by an unreleased build), record the version without running anything:

```bash
indexer migrate baseline 2 --db-host localhost --db-user postgres --db-name gnoland
```

This marks every migration up to and including that version as applied. Recording the wrong version makes
`migrate up` skip changes the database needs.

## Adding a migration

1. Change the structs in `pkgs/schema` so `setup create-db` creates the new schema.
2. Add `indexer/migrations/sql/NNNN_<release>.sql` that takes the previous schema to the new one. Don't add
   `BEGIN` / `COMMIT`, the runner wraps the file in a transaction. Statements that can't run inside a
   transaction (like `refresh_continuous_aggregate`) go in the migration's `After` list.
3. Append the migration to the list in `indexer/migrations/migrations.go` with the next version number.
   Use whole integer numbers for marking the versions.
4. Check that a database migrated from the previous release and one created with `setup create-db` end up
   with the same schema, for example by comparing `pg_dump --schema-only` of both.

Released migrations are never edited, a fix goes into a new migration.
