# Database data model and schema

This file will outline how the data is gathered, stored and processed by the indexer.

## Data flow from the RPC node

The indexer will connect to the RPC node and will start to gather the data from the node.
It collects and processes data by using batch processing. For live mode the indexer will collect the data up to the
latest block height and then it will process it in batches.

The data is gathered in the following way:

```mermaid
flowchart TD
    RPC[\RPC Node/]

    subgraph TX Pipeline
        FT([Fetch Transactions])
        D[Transactions]
        E[Transaction Messages]
        F[Gno Regular Addresses]
    end

    subgraph Validator Pipeline
        C{Validator Block Signings}
        G[Gno Validator Addresses]
    end

    BH[Block Height]
    I[(TimescaleDB)]

    RPC --> |Gather Block Height| BH
    RPC --> |Gather Validator Block Signings| C
    RPC --> FT
    BH --> |Gather TX Hashes| FT

    FT --> D
    D --> |Process Transactions| E
    E --> |Process Transaction Messages| F
    C --> |Process Validator Block Signings| G

    BH --> I
    D --> I
    E --> I
    F --> I
    C --> I
    G --> I
```

First the indexer gathers block data and validator signing for those blocks. If there are transactions the
tx hashes are gathered from the block height data and are queried from the RPC. At that moment the transaction
data is gathered and processes and all of the transaction general data and messages contained in the transaction
are stored in the database. The regular and validator addresses are processed in that way that the addresses are
stored as unique int32 ids and then referenced by the integer value in the transaction tables.

## Database schema

```mermaid
erDiagram
    blocks {
        BYTEA hash
        BIGINT height PK
        TIMESTAMPTZ timestamp PK
        INTEGER proposer
        chain_name chain_name
        BYTEA[] txs
    }
    validator_block_signings {
        BIGINT height FK
        TIMESTAMPTZ timestamp
        INTEGER proposer FK
        INTEGER[] signed_vals
        chain_name chain_name
    }
    transactions_general {
        BYTEA tx_hash PK
        chain_name chain_name PK
        TIMESTAMPTZ timestamp PK
        BIGINT block_height
        TEXT[] msg_types
        Event[] tx_events
        BYTEA tx_events_compressed
        BOOLEAN compression_on
        BIGINT gas_used
        BIGINT gas_wanted
        Amount fee
    }
    gno_addresses {
        INTEGER GENERATED ALWAYS AS IDENTITY id PK
        TEXT address UNIQUE
        chain_name chain_name UNIQUE
    }
    gno_validator_addresses {
        INTEGER GENERATED ALWAYS AS IDENTITY id PK
        TEXT address UNIQUE
        chain_name chain_name UNIQUE
    }
    address_tx {
        INTEGER address PK
        BYTEA tx_hash
        chain_name chain_name
        TIMESTAMPTZ timestamp PK
        TEXT[] msg_types
    }
    msg_send {
        BYTEA tx_hash PK
        SMALLINT message_counter PK
        TIMESTAMPTZ timestamp PK
        INTEGER from_address
        INTEGER to_address
        INTEGER[] signers
        Amount[] amount
    }
    msg_call {
        BYTEA tx_hash PK
        SMALLINT message_counter PK
        TIMESTAMPTZ timestamp PK
        INTEGER caller
        TEXT pkg_path
        TEXT func_name
        TEXT args
        INTEGER[] signers
        Amount[] send
        Amount[] max_deposit
    }
    msg_add_package {
        BYTEA tx_hash PK
        SMALLINT message_counter PK
        TIMESTAMPTZ timestamp PK
        INTEGER creator
        TEXT pkg_path
        TEXT pkg_name
        TEXT[] pkg_file_names
        INTEGER[] signers
        Amount[] send
        Amount[] max_deposit
    }
    msg_run {
        BYTEA tx_hash PK
        SMALLINT message_counter PK
        TIMESTAMPTZ timestamp PK
        INTEGER caller
        TEXT pkg_path
        TEXT pkg_name
        TEXT[] pkg_file_names
        INTEGER[] signers
        Amount[] send
        Amount[] max_deposit
    }

    block_counter {
        TIMESTAMPTZ time_bucket
        BIGINT block_count
        chain_name chain_name
    }
    tx_counter {
        TIMESTAMPTZ time_bucket
        BIGINT transaction_count
        chain_name chain_name
    }
    validator_signing_counter {
        TIMESTAMPTZ time_bucket
        BIGINT validator_id
        BIGINT blocks_signed
        chain_name chain_name
    }
    daily_active_accounts {
        TIMESTAMPTZ time_bucket
        BIGINT active_account_count
        chain_name chain_name
    }
    fee_volume {
        TIMESTAMPTZ time_bucket
        TEXT denom
        BIGINT volume
        chain_name chain_name
    }

    blocks ||--o{ transactions_general : "contains"
    blocks ||--o{ validator_block_signings : "has"
    gno_validator_addresses ||--o{ validator_block_signings : "signs"
    gno_validator_addresses ||--o{ blocks : "proposes"

    transactions_general ||--o{ address_tx : "involves"
    gno_addresses ||--o{ address_tx : "participates"

    transactions_general ||--o{ msg_send : "contains"
    transactions_general ||--o{ msg_call : "contains"
    transactions_general ||--o{ msg_add_package : "contains"
    transactions_general ||--o{ msg_run : "contains"

    gno_addresses ||--o{ msg_send : "from/to"
    gno_addresses ||--o{ msg_call : "caller"
    gno_addresses ||--o{ msg_add_package : "creator"
    gno_addresses ||--o{ msg_run : "caller"

    block_counter ||--o{ blocks : "count"
    tx_counter ||--o{ transactions_general : "count"
    validator_signing_counter ||--o{ validator_block_signings : "count"
    daily_active_accounts ||--o{ address_tx : "count"
    fee_volume ||--o{ transactions_general : "sum"
```

And custom types:

```mermaid
erDiagram
    amount {
        NUMERIC amount
        TEXT denom
    }
    attribute {
        TEXT key
        TEXT value
    }
    event {
        TEXT at_type
        TEXT type
        Attribute[] attributes
        TEXT pkg_path
    }
    transaction_general {
        Event[] tx_events
    }
    msg_send {
        Amount[] amount
    }
    msg_call {
        Amount[] send
        Amount[] max_deposit
    }
    msg_add_package {
        Amount[] send
        Amount[] max_deposit
    }
    msg_run {
        Amount[] send
        Amount[] max_deposit
    }

    amount ||--o{ transaction_general : "fee"
    amount ||--o{ msg_send : "amount"
    amount ||--o{ msg_call : "send"
    amount ||--o{ msg_add_package : "send"
    amount ||--o{ msg_add_package : "max_deposit"
    amount ||--o{ msg_run : "send"
    amount ||--o{ msg_run : "max_deposit"
    attribute ||--o{ event : "attributes"
    transaction_general ||--o{ event : "contains"
```
