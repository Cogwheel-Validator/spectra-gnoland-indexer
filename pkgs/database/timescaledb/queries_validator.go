package timescaledb

import (
	"context"
	"fmt"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
)

func (t *TimescaleDb) GetAllValidators(ctx context.Context, chainName string) (*database.ValidatorList, error) {
	query := `
	SELECT
	address
	FROM gno_validators
	WHERE chain_name = $1
	ORDER BY id ASC
	`
	var validators = make([]string, 0)
	rows, err := t.pool.Query(ctx, query, chainName)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var valAddr string
		err := rows.Scan(&valAddr)
		if err != nil {
			return nil, err
		}
		validators = append(validators, valAddr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(validators) == 0 {
		return nil, fmt.Errorf("validator list: %w", database.ErrNotFound)
	}

	return &database.ValidatorList{
		ValAddresses: validators,
	}, nil
}

func (t *TimescaleDb) GetValidatorLastNSigning(
	ctx context.Context,
	valAddr string,
	chainName string,
	limit uint64,
	orderBy database.SortOrder,
) (database.ValidatorSigningsForLastNBlocks, error) {
	result := make(database.ValidatorSigningsForLastNBlocks, 0, limit)

	query1 := `
	SELECT
	    id
	FROM
	    gno_validators
	WHERE
	    address = $1 AND chain_name = $2
	`

	var validatorId uint64
	err := t.pool.QueryRow(ctx, query1, valAddr, chainName).Scan(&validatorId)
	if err != nil {
		return nil, err
	}

	query2 := `
   	SELECT
        max(block_height)
    FROM
        validator_block_signing
    WHERE
        chain_name = $1
	`

	var maxBlockHeight uint64
	err = t.pool.QueryRow(ctx, query2, chainName).Scan(&maxBlockHeight)
	if err != nil {
		return nil, err
	}

	startHeight := maxBlockHeight - uint64(limit-1) // minus one because SQL BETWEEN is inclusive
	order := orderBy.SQL()

	query3 := fmt.Sprintf(`
	SELECT
        v.block_height AS height,
        COALESCE(v.signed_vals @> '{%d}', false) AS signed,
        COALESCE(gv.id = v.proposer, false) AS proposed
    FROM validator_block_signing v
    JOIN gno_validators gv ON gv.id = $1 AND gv.chain_name = $2
    WHERE v.block_height BETWEEN $3 AND $4 AND v.chain_name = $2
    ORDER BY height %s
	    `, validatorId, order)

	rows, err := t.pool.Query(ctx, query3, validatorId, chainName, startHeight, maxBlockHeight)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var height uint64
		var signed, proposed bool
		if err := rows.Scan(&height, &signed, &proposed); err != nil {
			return nil, err
		}
		result = append(result, database.ValInfoPerBlock{
			Height:   height,
			Signed:   signed,
			Proposed: proposed,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
