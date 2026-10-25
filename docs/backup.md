# Backing up and restoring the indexer database

The indexer binary has a `backup` command group. It talks to Postgres over its
normal connection only; no `pg_dump`, `psql` or other client tools are needed.
All SQL it runs lives in `indexer/backup/sql/` and is embedded in the binary.

| command | does |
|---|---|
| `indexer backup create` | backs up a range of block heights of one chain |
| `indexer backup restore --target-db DB` | builds `DB` from the backups, or brings an earlier restore up to date |
| `indexer backup verify --target-db DB` | compares the restore with the live source, row for row |
| `indexer backup list` | lists the backups in the backup directory |

## Quick start

```bash
export DB_HOST=localhost DB_PORT=6543 DB_USER=postgres DB_PASSWORD=...   # or the --db-* flags

./build/indexer backup create --dir /srv/gnoland-backups          # first run: from the chain's first block
./build/indexer backup create --dir /srv/gnoland-backups          # later runs: continue where the last one ended

./build/indexer backup restore --dir /srv/gnoland-backups --target-db gnoland_restored
./build/indexer backup verify  --target-db gnoland_restored       # exits non-zero on any mismatch
./build/indexer backup list    --dir /srv/gnoland-backups
```

The connection flags (`--db-host`, `--db-port`, `--db-user`, `--ssl-mode`, ...)
and the `DB_*` environment variables work as for `indexer setup`. The password
comes from `DB_PASSWORD` or an interactive prompt. `--db-name` is the source
database (default `gnoland`). The database user must be able to read every
table; restore needs a user that can create databases and extensions (the
`postgres` account used for `setup create-db`).

## Choosing the range

Every backup covers one chain over an inclusive range of block heights.

| flags | range |
|---|---|
| none | from right after the previous backup in `--dir` (or the chain's first block) up to the last height before the first indexing gap |
| `--from-height N --to-height M` | exactly heights `N..M` |
| `--from-height N` | `N` up to the last height before the first gap |
| `--to-height M` | from right after the previous backup up to `M` |
| `--from-time T` / `--to-time T` | instead of a height: the first block at or after `T` / the last block before `T` (RFC 3339) |
| `--chain-name C` | which chain; only needed when the database holds more than one |

A height counts as indexed once it has both its `blocks` and its
`validator_block_signing` row (see [How the range is checked](#how-the-range-is-checked)).
An explicit range with missing heights is refused unless `--allow-gaps` is
given; the number of missing heights is then recorded in the manifest.

Running `backup create` with no range flags on a schedule gives a chain of
incremental backups. When nothing new has been indexed it says so and exits 0:

```cron
0 */3 * * *  . /etc/gnoland-backup.env && /opt/gnoland/indexer backup create --dir /srv/gnoland-backups
```

Only one backup per chain runs at a time (a lock file in the chain's backup
directory); an overlapping run fails instead of waiting.

## Why not plain `pg_dump`

`pg_dump` of a TimescaleDB database dumps the chunk tables and the
`_timescaledb_catalog` rows that describe them. Compressed chunks come out as
opaque compressed batches. The result only restores correctly into the **same
TimescaleDB version**, wrapped in `timescaledb_pre_restore()` /
`timescaledb_post_restore()`, and it can't be cut at a block height.

The backup command never touches Timescale internals:

- **Data** is exported with `COPY (SELECT ... FROM hypertable WHERE ...)`.
  Selecting from a hypertable decompresses compressed chunks on the fly, so the
  backup is plain rows whatever compression state the source is in.
- **Schema** is regenerated from the source catalog by the queries in
  `indexer/backup/sql/schema/`: enum and composite types, tables, constraints,
  indexes, `create_hypertable(...)` with the chunk interval, compression
  settings and policies, grants, continuous aggregate definitions and their
  policies. Each query returns `(key, statement)` rows.
- **Restore** runs the statements whose key the target does not have yet, so
  the same files build an empty database and add the tables, columns and enum
  values the source gained since an earlier restore. TimescaleDB builds its
  own chunks as the rows are `COPY`'d back in. `--compress` runs the
  compression policies right away; otherwise the background jobs do it later.

## What one backup contains

```
<dir>/<chain>/<id>/               id = 20261019T125227Z-1-511821  (creation time, from height, to height)
  manifest.json   chain, height range and its time window, snapshot time, address marks,
                  every data file with its columns, row count and sha256, source versions
  schema.json     the DDL statements, with their keys and phase
  data/NNN-<table>.copy.zst   COPY text format, zstd compressed
  DONE            sha256 of manifest.json; written last
```

A backup directory only gets its final name after everything is written and
checksummed. A failed run deletes its `.partial` directory. Directories and
files are created owner-only: `api_keys` (key hashes) is in every backup.

## How the data is split

Every read of one backup goes through a single exported snapshot
(`pg_export_snapshot()`), so all tables are seen at the same instant, also with
`--jobs N` parallel COPY sessions. Before the snapshot is taken, the backup
takes `ACCESS SHARE` locks on every public table and hypertable. That doesn't
block the indexer's inserts. It does stop the compression policy from
`TRUNCATE`-ing a chunk halfway through the export, which is the one thing MVCC
does not protect against.

Each public table falls into one of three classes (decided from the catalog on
every run by `sql/tables.sql`, so new message tables are picked up
automatically):

| class | tables | per backup |
|---|---|---|
| `chain_ts` | hypertables with `chain_name` (`blocks`, `validator_block_signing`, `transaction_general`, `address_tx`, every `*_msg_*` table) | the chain's rows inside the backup's time window |
| `idseq` | plain tables with an identity column (`gno_addresses`, `gno_validators`) | rows with `id >` the address floor |
| `snapshot` | other plain tables (`api_keys`, `schema_migrations`) | the whole table |

### From heights to a time window

Message tables carry their block's timestamp but no height. Block times
strictly increase, so heights `N..M` are exactly the rows with
`timestamp(N) <= timestamp <= timestamp(M)`. `sql/bounds.sql` checks that the
number of blocks by height and by time agree, and the backup refuses the range
if they don't.

### How the range is checked

The orchestrator commits every chunk in two phases
(`indexer/orchestrator/operator.go`):

1. validator addresses, transactions, regular addresses, `address_tx` and the
   message tables
2. after phase 1 has finished: `blocks` and `validator_block_signing`,
   concurrently with each other

So a height with **both** a `blocks` row and a `validator_block_signing` row
is fully committed (`sql/progress.sql`). An automatic range stops before the
first height that is not, so a hole that gets filled later (for example by a
`historic` re-run) is still ahead of the next backup. If the two phases ever
change order, this rule has to be revisited.

### Addresses

Address ids come from one identity per table and are referenced from every
message table. The indexer inserts addresses in their own autocommit `COPY`
before the rows that use them, so every address the backed-up rows reference
is visible in the same snapshot.

A backup that continues exactly where the previous one ended exports only
`id >` the previous backup's mark. The mark is the highest id seen
`--addr-settle` (default 10 seconds) **before** the snapshot, so an insert still
in flight when a snapshot opens always has an id above it. Some addresses
therefore show up in two consecutive backups, which is harmless: restore
inserts them with `ON CONFLICT (id) DO NOTHING`. Any other range (the first
backup, an explicit older range) exports the whole address tables. Restore
refuses a continuing backup whose predecessor the target does not hold, and
moves the identity sequences to at least the source's value.

### Late rows

Restore **replaces** each backup's window: it deletes the chain's rows in that
time window and loads the backup's rows, in one transaction. If rows arrive in
a range that was already backed up (a `historic` re-run, a retried chunk),
back that range up again and restore it:

```bash
indexer backup create --from-height 150000 --to-height 250000
indexer backup restore --target-db gnoland_restored
```

### Continuous aggregates

Their data is never backed up. Restore creates missing aggregates empty and
refreshes them over the whole range; aggregates that already exist are
refreshed over the windows of the backups just applied (widened by one bucket
on each side). Only then are the refresh and compression policies attached, so
they can't collide with that refresh. `--no-refresh` skips the refresh.

## Restoring

`restore` applies every backup in `--dir/<chain>` that the target does not
hold yet, oldest first (`--up-to ID` stops earlier). Before it creates the
database or writes any data, it checks every data file against its checksum
and that the address exports build on each other. Each backup's data lands in one transaction. The target
records what it holds in `gnoland_backup.applied`, so re-running the same
command later applies only the new backups. Restore refuses to write into an
existing database that is neither empty nor one of its own restores.

Onto a new server:

1. Install PostgreSQL and TimescaleDB (+ timescaledb_toolkit). The logical
   format doesn't need the exact TimescaleDB version.
2. Copy the backup directory.
3. `indexer backup restore --dir <dir> --target-db gnoland` with that server's
   connection flags.
4. Create the application users with `indexer setup create-user`; grants to
   roles that did not exist on the new server were skipped with a warning.
5. Point the indexer and API at the database. Live mode resumes from the
   highest block in `blocks`.

## Verifying a restore

`verify` reads the ranges the target holds from `gnoland_backup.applied`,
merges them into runs of consecutive heights, and compares each with the
source (`--db-name`):

- every per-chain table: row count plus an order-independent hash of every row
  in the run's time window
- address tables: identical up to the highest restored mark, and every row
  above it identical in the source
- every address id referenced by restored rows resolves (`sql/refs.sql`: every
  `integer`/`integer[]` column of the per-chain hypertables;
  `validator_block_signing` points at `gno_validators`, the rest at
  `gno_addresses`)
- continuous aggregates: every bucket that lies completely inside a run matches
- schema: every key produced by the schema queries, compared statement by
  statement

The whole-table snapshots (`api_keys`, `schema_migrations`), grants, and objects
the source gained after the backup are reported as `WARN`. Anything else is a
`FAIL`, and the command exits non-zero. `-v` also prints the passing checks.

## Assumptions and limits

- `blocks` + `validator_block_signing` and the `chain_name` enum are the
  progress marker. Block times must strictly increase within a chain.
- A hypertable without a `NOT NULL chain_name` column is refused, since the
  backup would not know how to split it.
- The locks mean an `ALTER TABLE` (a migration) waits until the backup finishes.
- Roles and their passwords are not backed up.
- Each backup is applied in a single transaction. That's fine at the current
  size; very large ranges need correspondingly more time and WAL. Split them
  with `--to-height`.
- Retention: deleting a backup is safe for the restores that already hold it.
  To restore from scratch, keep a chain that starts with a backup of whole
  address tables (the first backup, or any explicit range) and every
  continuation after it.
