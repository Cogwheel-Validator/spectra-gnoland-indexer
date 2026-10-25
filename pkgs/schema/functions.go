package schema

import (
	"fmt"

	dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/db_init"
)

// resolveFunctionSQL builds a function that returns the id of every given address,
// inserting the ones the table does not have yet. Each address seen on the chain
// for the first time has the chain added to its chains column and is counted in
// chain_stats.
//
// Every statement in a plpgsql function takes a fresh snapshot, so the final
// SELECT also sees addresses a concurrent caller committed meanwhile.
func resolveFunctionSQL(name, table, countCol string) string {
	return fmt.Sprintf(`
CREATE OR REPLACE FUNCTION %[1]s(p_chain chain_name, p_addresses text[])
RETURNS TABLE (address text, id integer)
LANGUAGE plpgsql AS $$
#variable_conflict use_column
DECLARE
    inserted bigint;
    marked   bigint;
BEGIN
    INSERT INTO %[2]s (address, chains)
    SELECT DISTINCT i.address, ARRAY[p_chain]
    FROM unnest(p_addresses) AS i(address)
    WHERE NOT EXISTS (SELECT 1 FROM %[2]s t WHERE t.address = i.address)
    ORDER BY 1
    ON CONFLICT (address) DO NOTHING;
    GET DIAGNOSTICS inserted = ROW_COUNT;

    PERFORM 1 FROM %[2]s t
    WHERE t.address = ANY(p_addresses) AND NOT (p_chain = ANY(t.chains))
    ORDER BY t.id
    FOR UPDATE;

    UPDATE %[2]s t
    SET chains = array_append(t.chains, p_chain)
    WHERE t.address = ANY(p_addresses) AND NOT (p_chain = ANY(t.chains));
    GET DIAGNOSTICS marked = ROW_COUNT;

    IF inserted + marked > 0 THEN
        INSERT INTO chain_stats (chain_name, %[3]s)
        VALUES (p_chain, inserted + marked)
        ON CONFLICT (chain_name) DO UPDATE SET %[3]s = chain_stats.%[3]s + EXCLUDED.%[3]s;
    END IF;

    RETURN QUERY
    SELECT t.address, t.id FROM %[2]s t WHERE t.address = ANY(p_addresses);
END
$$;`, name, table, countCol)
}

// Functions returns every database function the indexer calls.
func Functions() []dbinit.FunctionDef {
	return []dbinit.FunctionDef{
		{Name: "resolve_addresses", SQL: resolveFunctionSQL("resolve_addresses", "gno_addresses", "address_count")},
		{Name: "resolve_validators", SQL: resolveFunctionSQL("resolve_validators", "gno_validators", "validator_count")},
	}
}
