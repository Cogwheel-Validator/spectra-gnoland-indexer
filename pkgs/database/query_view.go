package database

import (
	"context"
	"fmt"
	"time"
)

func (t *TimescaleDb) GetBlockCountByDate(
	ctx context.Context,
	chainName string,
	date1 time.Time,
	date2 time.Time,
) ([]*BlockCountByDate, error) {
	date1 = time.Date(date1.Year(), date1.Month(), date1.Day(), 0, 0, 0, 0, date1.Location())
	date2 = time.Date(date2.Year(), date2.Month(), date2.Day(), 23, 59, 59, 0, date2.Location())

	query := `
	SELECT
	time_bucket('1 day', time_bucket)::date as date, 
	SUM(block_count) as block_count
	FROM block_counter
	WHERE chain_name = $1
	AND time_bucket >= $2 AND time_bucket <= $3
	GROUP BY time_bucket('1 day', time_bucket)
	ORDER BY date DESC
	`

	rows, err := t.pool.Query(ctx, query, chainName, date1, date2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var blockCountByDates []*BlockCountByDate
	for rows.Next() {
		blockCountByDate := &BlockCountByDate{}
		err := rows.Scan(&blockCountByDate.Date, &blockCountByDate.Count)
		if err != nil {
			return nil, err
		}
		blockCountByDates = append(blockCountByDates, blockCountByDate)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return blockCountByDates, nil
}

func (t *TimescaleDb) GetBlockCount24h(
	ctx context.Context,
	chainName string,
) (int64, error) {
	query := `
	SELECT
	SUM(block_count) as block_count
	FROM block_counter
	WHERE chain_name = $1
	AND time_bucket >= NOW() - INTERVAL '24 hours'
	AND time_bucket < NOW()
	`
	row := t.pool.QueryRow(ctx, query, chainName)
	var blockCount int64
	err := row.Scan(&blockCount)
	if err != nil {
		return 0, err
	}
	return blockCount, nil
}

func (t *TimescaleDb) GetDailyActiveAccount(
	ctx context.Context,
	chainName string,
	date1 time.Time,
	date2 time.Time,
) ([]*DailyActiveAccount, error) {
	query := `
	SELECT
	time_bucket::date as date,
	active_account_count as count
	FROM daily_active_accounts
	WHERE chain_name = $1
	AND time_bucket >= $2 AND time_bucket <= $3
	ORDER BY date DESC
	`
	rows, err := t.pool.Query(ctx, query, chainName, date1, date2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dailyActiveAccounts []*DailyActiveAccount
	for rows.Next() {
		dailyActiveAccount := &DailyActiveAccount{}
		err := rows.Scan(&dailyActiveAccount.Date, &dailyActiveAccount.Count)
		if err != nil {
			return nil, err
		}
		dailyActiveAccounts = append(dailyActiveAccounts, dailyActiveAccount)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dailyActiveAccounts, nil
}

func (t *TimescaleDb) GetTotalTxCount(
	ctx context.Context,
	chainName string,
) (int64, error) {
	query := `
	SELECT
	SUM(transaction_count) as total_tx_count
	FROM tx_counter
	WHERE
	chain_name = $1
	`
	row := t.pool.QueryRow(ctx, query, chainName)
	var totalTxCount int64
	err := row.Scan(&totalTxCount)
	if err != nil {
		return 0, err
	}
	return totalTxCount, nil
}

func (t *TimescaleDb) GetTotalTxCount24h(
	ctx context.Context,
	chainName string,
) (int64, error) {
	query := `
	SELECT
	SUM(transaction_count) as tx_count_24h
	FROM tx_counter
	WHERE
	chain_name = $1
	AND time_bucket >= NOW() - INTERVAL '24 hours'
	AND time_bucket < NOW()
	`
	row := t.pool.QueryRow(ctx, query, chainName)
	var txCount24h int64
	err := row.Scan(&txCount24h)
	if err != nil {
		return 0, err
	}
	return txCount24h, nil
}

func (t *TimescaleDb) GetTotalTxCountByDate(
	ctx context.Context,
	chainName string,
	date1 time.Time,
	date2 time.Time,
) ([]*TxCountTimeRange, error) {
	date1 = time.Date(date1.Year(), date1.Month(), date1.Day(), 0, 0, 0, 0, date1.Location())
	date2 = time.Date(date2.Year(), date2.Month(), date2.Day(), 23, 59, 59, 0, date2.Location())

	query := `
	SELECT
	time_bucket_gapfill('1 day', time_bucket)::date as date,
	coalesce(SUM(transaction_count), 0) as tx_count
	FROM tx_counter
	WHERE
	chain_name = $1
	AND time_bucket >= $2 AND time_bucket <= $3
	GROUP BY time_bucket('1 day', time_bucket)
	ORDER BY date DESC
	`
	rows, err := t.pool.Query(ctx, query, chainName, date1, date2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var txCountTimeRanges []*TxCountTimeRange
	for rows.Next() {
		txCountTimeRange := &TxCountTimeRange{}
		err := rows.Scan(&txCountTimeRange.Time, &txCountTimeRange.Count)
		if err != nil {
			return nil, err
		}
		txCountTimeRanges = append(txCountTimeRanges, txCountTimeRange)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return txCountTimeRanges, nil
}

func (t *TimescaleDb) GetTotalTxCountByHour(
	ctx context.Context,
	chainName string,
	date1 time.Time,
	date2 time.Time,
) ([]*TxCountTimeRange, error) {
	query := `
	SELECT
	time_bucket_gapfill('1 hour', time_bucket) as timestamp,
	coalesce(SUM(transaction_count), 0) as tx_count
	FROM tx_counter
	WHERE
	chain_name = $1
	AND time_bucket >= $2 AND time_bucket <= $3
	GROUP BY time_bucket
	ORDER BY timestamp DESC
	`
	rows, err := t.pool.Query(ctx, query, chainName, date1, date2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var txCountTimeRanges []*TxCountTimeRange
	for rows.Next() {
		txCountTimeRange := &TxCountTimeRange{}
		err := rows.Scan(&txCountTimeRange.Time, &txCountTimeRange.Count)
		if err != nil {
			return nil, err
		}
		txCountTimeRanges = append(txCountTimeRanges, txCountTimeRange)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return txCountTimeRanges, nil
}

func (t *TimescaleDb) GetVolumeByDate(
	ctx context.Context,
	chainName string,
	date1 time.Time,
	date2 time.Time,
) (VolumeByDenom, error) {
	date1 = time.Date(date1.Year(), date1.Month(), date1.Day(), 0, 0, 0, 0, date1.Location())
	date2 = time.Date(date2.Year(), date2.Month(), date2.Day(), 23, 59, 59, 0, date2.Location())

	query := `
	SELECT
	time_bucket_gapfill('1 day', time_bucket)::date as date,
	coalesce(SUM(volume), 0) as volume,
	denom
	FROM fee_volume
	WHERE
	chain_name = $1
	AND time_bucket >= $2 AND time_bucket <= $3
	GROUP BY time_bucket('1 day', time_bucket), denom
	ORDER BY date DESC
	`
	rows, err := t.pool.Query(ctx, query, chainName, date1, date2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var feeVolumeTimeRanges = make(VolumeByDenom)
	for rows.Next() {
		denomVolume := &DenomVolume{}
		denom := ""
		err := rows.Scan(&denomVolume.Time, &denomVolume.Volume, &denom)
		if err != nil {
			return nil, err
		}
		feeVolumeTimeRanges[denom] = append(feeVolumeTimeRanges[denom], denomVolume)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return feeVolumeTimeRanges, nil
}

func (t *TimescaleDb) GetVolumeByHour(
	ctx context.Context,
	chainName string,
	date1 time.Time,
	date2 time.Time,
) (VolumeByDenom, error) {
	query := `
	SELECT
	time_bucket_gapfill('1 hour', time_bucket) as time,
	coalesce(SUM(volume), 0) as volume,
	denom
	FROM fee_volume
	WHERE
	chain_name = $1
	AND time_bucket >= $2 AND time_bucket <= $3
	GROUP BY time_bucket('1 hour', time_bucket), denom
	ORDER BY time DESC
	`
	rows, err := t.pool.Query(ctx, query, chainName, date1, date2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var feeVolumeTimeRanges = make(VolumeByDenom)
	for rows.Next() {
		denomVolume := &DenomVolume{}
		denom := ""
		err := rows.Scan(&denomVolume.Time, &denomVolume.Volume, &denom)
		if err != nil {
			return nil, err
		}
		feeVolumeTimeRanges[denom] = append(feeVolumeTimeRanges[denom], denomVolume)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return feeVolumeTimeRanges, nil
}

func (t *TimescaleDb) GetValidatorSigning24h(
	ctx context.Context,
	validatorAddress string,
	chainName string,
) (*ValidatorSigning, error) {
	// check if validator address exists and get it's id
	query1 := `
	SELECT
	id
	FROM gno_validators
	WHERE address = $1
	AND chain_name = $2
	`
	row1 := t.pool.QueryRow(ctx, query1, validatorAddress, chainName)
	var validatorId int32
	err := row1.Scan(&validatorId)
	if err != nil {
		return nil, fmt.Errorf("validator seems to not exist: %w", err)
	}

	query2 := `
	SELECT
    coalesce(sum(vsc.blocks_signed), 0) AS blocks_signed,
    coalesce(sum(bc.block_count), 0) - coalesce(sum(vsc.blocks_signed), 0) AS blocks_not_signed,
    coalesce(sum(bc.block_count), 0) AS total_blocks,
    coalesce(round(coalesce(sum(vsc.blocks_signed), 0)::numeric / nullif(sum(bc.block_count), 0) * 100, 2), 0) AS signing_rate_pct
	FROM validator_signing_counter vsc
	LEFT JOIN block_counter bc
		ON  bc.time_bucket = vsc.time_bucket
		AND bc.chain_name  = vsc.chain_name
	WHERE vsc.chain_name    = $1
		AND vsc.validator_id  = $2
		AND vsc.time_bucket >= now() - INTERVAL '24 hours'
		AND vsc.time_bucket <  now();
	`

	row := t.pool.QueryRow(ctx, query2, chainName, validatorId)
	var validatorSigning ValidatorSigning
	err = row.Scan(
		&validatorSigning.BlocksSigned,
		&validatorSigning.BlocksMissed,
		&validatorSigning.TotalBlocks,
		&validatorSigning.SigningRate,
	)
	if err != nil {
		return nil, err
	}
	return &validatorSigning, nil
}

func (t *TimescaleDb) GetValidatorSigningByHour(
	ctx context.Context,
	validatorAddress string,
	chainName string,
	date1 time.Time,
	date2 time.Time,
) ([]*ValidatorSigning, error) {
	// check if validator address exists and get it's id
	query1 := `
	SELECT
	id
	FROM gno_validators
	WHERE address = $1
	AND chain_name = $2
	`
	row1 := t.pool.QueryRow(ctx, query1, validatorAddress, chainName)
	var validatorId int32
	err := row1.Scan(&validatorId)
	if err != nil {
		return nil, fmt.Errorf("validator seems to not exist: %w", err)
	}

	query2 := `
	SELECT
	time_bucket_gapfill('1 hour', vsc.time_bucket) as time,
	coalesce(sum(vsc.blocks_signed), 0) as blocks_signed,
	coalesce(sum(bc.block_count), 0) - coalesce(sum(vsc.blocks_signed), 0) as blocks_not_signed,
	coalesce(sum(bc.block_count), 0) as total_blocks,
	coalesce(round(coalesce(sum(vsc.blocks_signed), 0)::numeric / nullif(sum(bc.block_count), 0) * 100, 2), 0) as signing_rate_pct
	FROM validator_signing_counter vsc
	LEFT JOIN block_counter bc
		ON  bc.time_bucket = vsc.time_bucket
		AND bc.chain_name  = vsc.chain_name
	WHERE vsc.chain_name    = $1
		AND vsc.validator_id  = $2
		AND vsc.time_bucket >= $3 AND vsc.time_bucket <= $4
	GROUP BY time_bucket('1 hour', vsc.time_bucket)
	ORDER BY time DESC
	`
	rows, err := t.pool.Query(ctx, query2, chainName, validatorId, date1, date2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var validatorSignings []*ValidatorSigning
	for rows.Next() {
		validatorSigning := &ValidatorSigning{}
		err := rows.Scan(&validatorSigning.Time, &validatorSigning.BlocksSigned, &validatorSigning.BlocksMissed, &validatorSigning.TotalBlocks, &validatorSigning.SigningRate)
		if err != nil {
			return nil, err
		}
		validatorSignings = append(validatorSignings, validatorSigning)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return validatorSignings, nil
}
