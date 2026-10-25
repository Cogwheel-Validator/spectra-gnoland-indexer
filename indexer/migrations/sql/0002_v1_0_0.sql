-- v0.10.1 -> v1.0.0
--
-- 1. gno_addresses and gno_validators keep one row per address shared by every
--    chain: chain_name is replaced by a chains array and the unique key is the
--    address alone.
-- 2. chain_stats keeps running address and validator counts per chain.
-- 3. resolve_addresses / resolve_validators look up and insert address ids.
-- 4. tx_counter gains failed_count, gas_used and gas_wanted. Continuous
--    aggregates can't have columns added in place, so the view is dropped and
--    recreated with the settings "setup create-db" applies.
--
-- Grants other roles hold on gno_addresses are copied to chain_stats, and the
-- grants on tx_counter are restored after it is recreated, so existing reader
-- and writer users keep working.

-- An address indexed on two chains has two ids in v0.10.1. Merging them would
-- mean rewriting every hypertable row that references the ids, so it will refuse.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM gno_addresses GROUP BY address HAVING count(*) > 1) THEN
        RAISE EXCEPTION 'gno_addresses holds the same address under more than one chain'
            USING HINT = 'This migration can only convert databases with distinct addresses. '
                'Create a new database with "setup create-db" and re-index.';
    END IF;
    IF EXISTS (SELECT 1 FROM gno_validators GROUP BY address HAVING count(*) > 1) THEN
        RAISE EXCEPTION 'gno_validators holds the same address under more than one chain'
            USING HINT = 'This migration can only convert databases with distinct addresses. '
                'Create a new database with "setup create-db" and re-index.';
    END IF;
END
$$;

-- Dropping chain_name also drops the UNIQUE (address, chain_name) constraint.
ALTER TABLE gno_addresses ADD COLUMN chains chain_name[] NOT NULL DEFAULT '{}';
UPDATE gno_addresses SET chains = ARRAY[chain_name];
ALTER TABLE gno_addresses DROP COLUMN chain_name;
ALTER TABLE gno_addresses ADD UNIQUE (address);

ALTER TABLE gno_validators ADD COLUMN chains chain_name[] NOT NULL DEFAULT '{}';
UPDATE gno_validators SET chains = ARRAY[chain_name];
ALTER TABLE gno_validators DROP COLUMN chain_name;
ALTER TABLE gno_validators ADD UNIQUE (address);

CREATE TABLE chain_stats (
    chain_name chain_name NOT NULL,
    address_count BIGINT DEFAULT 0 NOT NULL,
    validator_count BIGINT DEFAULT 0 NOT NULL,
    PRIMARY KEY (chain_name)
);

INSERT INTO chain_stats (chain_name, address_count, validator_count)
SELECT chain, sum(addresses), sum(validators)
FROM (
    SELECT unnest(chains) AS chain, 1 AS addresses, 0 AS validators FROM gno_addresses
    UNION ALL
    SELECT unnest(chains), 0, 1 FROM gno_validators
) counts
GROUP BY chain;

DO $$
DECLARE
    g record;
BEGIN
    FOR g IN
        SELECT a.grantee::regrole::text AS grantee, string_agg(a.privilege_type, ', ') AS privs
        FROM pg_class c, aclexplode(c.relacl) a
        WHERE c.oid = 'gno_addresses'::regclass AND a.grantee <> c.relowner AND a.grantee <> 0
        GROUP BY a.grantee
    LOOP
        EXECUTE format('GRANT %s ON TABLE chain_stats TO %s', g.privs, g.grantee);
    END LOOP;
END
$$;

CREATE OR REPLACE FUNCTION resolve_addresses(p_chain chain_name, p_addresses text[])
RETURNS TABLE (address text, id integer)
LANGUAGE plpgsql AS $$
#variable_conflict use_column
DECLARE
    inserted bigint;
    marked   bigint;
BEGIN
    INSERT INTO gno_addresses (address, chains)
    SELECT DISTINCT i.address, ARRAY[p_chain]
    FROM unnest(p_addresses) AS i(address)
    WHERE NOT EXISTS (SELECT 1 FROM gno_addresses t WHERE t.address = i.address)
    ORDER BY 1
    ON CONFLICT (address) DO NOTHING;
    GET DIAGNOSTICS inserted = ROW_COUNT;

    PERFORM 1 FROM gno_addresses t
    WHERE t.address = ANY(p_addresses) AND NOT (p_chain = ANY(t.chains))
    ORDER BY t.id
    FOR UPDATE;

    UPDATE gno_addresses t
    SET chains = array_append(t.chains, p_chain)
    WHERE t.address = ANY(p_addresses) AND NOT (p_chain = ANY(t.chains));
    GET DIAGNOSTICS marked = ROW_COUNT;

    IF inserted + marked > 0 THEN
        INSERT INTO chain_stats (chain_name, address_count)
        VALUES (p_chain, inserted + marked)
        ON CONFLICT (chain_name) DO UPDATE SET address_count = chain_stats.address_count + EXCLUDED.address_count;
    END IF;

    RETURN QUERY
    SELECT t.address, t.id FROM gno_addresses t WHERE t.address = ANY(p_addresses);
END
$$;

CREATE OR REPLACE FUNCTION resolve_validators(p_chain chain_name, p_addresses text[])
RETURNS TABLE (address text, id integer)
LANGUAGE plpgsql AS $$
#variable_conflict use_column
DECLARE
    inserted bigint;
    marked   bigint;
BEGIN
    INSERT INTO gno_validators (address, chains)
    SELECT DISTINCT i.address, ARRAY[p_chain]
    FROM unnest(p_addresses) AS i(address)
    WHERE NOT EXISTS (SELECT 1 FROM gno_validators t WHERE t.address = i.address)
    ORDER BY 1
    ON CONFLICT (address) DO NOTHING;
    GET DIAGNOSTICS inserted = ROW_COUNT;

    PERFORM 1 FROM gno_validators t
    WHERE t.address = ANY(p_addresses) AND NOT (p_chain = ANY(t.chains))
    ORDER BY t.id
    FOR UPDATE;

    UPDATE gno_validators t
    SET chains = array_append(t.chains, p_chain)
    WHERE t.address = ANY(p_addresses) AND NOT (p_chain = ANY(t.chains));
    GET DIAGNOSTICS marked = ROW_COUNT;

    IF inserted + marked > 0 THEN
        INSERT INTO chain_stats (chain_name, validator_count)
        VALUES (p_chain, inserted + marked)
        ON CONFLICT (chain_name) DO UPDATE SET validator_count = chain_stats.validator_count + EXCLUDED.validator_count;
    END IF;

    RETURN QUERY
    SELECT t.address, t.id FROM gno_validators t WHERE t.address = ANY(p_addresses);
END
$$;

CREATE TEMP TABLE tx_counter_grants ON COMMIT DROP AS
SELECT a.grantee::regrole::text AS grantee, string_agg(a.privilege_type, ', ') AS privs
FROM pg_class c, aclexplode(c.relacl) a
WHERE c.oid = 'tx_counter'::regclass AND a.grantee <> c.relowner AND a.grantee <> 0
GROUP BY a.grantee;

DROP MATERIALIZED VIEW tx_counter;

CREATE MATERIALIZED VIEW tx_counter
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', timestamp) AS time_bucket,
    chain_name,
    count(*) AS transaction_count,
    count(*) FILTER (WHERE NOT success) AS failed_count,
    sum(gas_used) AS gas_used,
    sum(gas_wanted) AS gas_wanted
FROM transaction_general
GROUP BY 1, 2
WITH NO DATA;

ALTER MATERIALIZED VIEW tx_counter SET (
    timescaledb.enable_columnstore = TRUE,
    timescaledb.segmentby = 'chain_name',
    timescaledb.orderby = 'time_bucket DESC'
);

SELECT add_continuous_aggregate_policy('tx_counter',
    start_offset => NULL,
    end_offset => INTERVAL '15 seconds',
    schedule_interval => INTERVAL '15 seconds'
);

CALL add_columnstore_policy('tx_counter', INTERVAL '1 month');

ALTER MATERIALIZED VIEW tx_counter SET (timescaledb.materialized_only = FALSE);

DO $$
DECLARE
    g record;
BEGIN
    FOR g IN SELECT grantee, privs FROM tx_counter_grants LOOP
        EXECUTE format('GRANT %s ON TABLE tx_counter TO %s', g.privs, g.grantee);
    END LOOP;
END
$$;
