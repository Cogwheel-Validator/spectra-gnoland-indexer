package database

import (
	"context"
	"fmt"
)

type KeyParams struct {
	RpmLimit int
	Name     string
	Prefix   string
	Hash     [32]byte
}

func (t *TimescaleDb) InsertApiKey(
	ctx context.Context, params KeyParams) error {
	result, err := t.pool.Exec(ctx, `
		INSERT INTO api_keys (prefix, hash, name, rpm_limit)
		VALUES ($1, $2, $3, $4)
		`, params.Prefix, params.Hash, params.Name, params.RpmLimit)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (t *TimescaleDb) GetAllApiKeys(ctx context.Context) ([][32]byte, error) {
	query := `
		SELECT hash
		FROM api_keys
		WHERE is_active = true
		`
	rows, err := t.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apiKeys := make([][32]byte, 0)
	for rows.Next() {
		var hash [32]byte
		err := rows.Scan(&hash)
		if err != nil {
			return nil, err
		}
		apiKeys = append(apiKeys, hash)
	}
	return apiKeys, nil
}

func (t *TimescaleDb) DisableKey(ctx context.Context, hash [32]byte) error {
	result, err := t.pool.Exec(ctx, `
		UPDATE api_keys
		SET is_active = false
		WHERE hash = $1
		`, hash)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (t *TimescaleDb) EnableKey(ctx context.Context, hash [32]byte) error {
	result, err := t.pool.Exec(ctx, `
		UPDATE api_keys
		SET is_active = true
		WHERE hash = $1
		`, hash)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (t *TimescaleDb) AdjustRpmLimit(
	ctx context.Context, hash [32]byte, rpmLimit int,
) error {
	result, err := t.pool.Exec(ctx, `
		UPDATE api_keys
		SET rpm_limit = $1
		WHERE hash = $2
		`, rpmLimit, hash)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}
