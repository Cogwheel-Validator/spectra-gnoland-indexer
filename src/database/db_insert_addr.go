package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// InsertAddresses inserts a slice of addresses into the database
//
// This is a method to insert a slice of addresses into the database
// It should preform better than using INSERT INTO... for a large number of addresses
// because it uses the COPY FROM command
//
// Usage:
//
// # Used inside of the address cache package to insert the addresses to the database
//
// Parameters:
//   - addresses: a slice of addresses to insert
//   - chainName: the name of the chain to insert the addresses to
//   - insertValidators: a boolean to indicate if the addresses are validators or accounts
//
// Returns:
//   - error: an error if the insertion fails
func (t *TimescaleDb) InsertAddresses(addresses []string, chainName string, insertValidators bool) error {
	column_names := []string{"address", "chain_name"}
	var table_name string
	if insertValidators {
		table_name = "gno_validators"
	} else {
		table_name = "gno_addresses"
	}
	// create interface to copy from slice to the db
	copy_from_slice := pgx.CopyFromSlice(len(addresses), func(i int) ([]any, error) {
		return []any{addresses[i], chainName}, nil
	})
	// copy the addresses to the db
	_, err := t.pool.CopyFrom(context.Background(), pgx.Identifier{table_name}, column_names, copy_from_slice)
	return err
}

// InsertAnAddress inserts a single address into the database
//
// This is a method to insert a single address into the database
// Although it might not be used as much it will serve as a safety net in case slice
// insert fails for some reason
//
// Usage:
//
// Used inside of the address cache package to insert a single address to the database
// only to be used in special cases
//
// Parameters:
//   - address: the address to insert
//   - chainName: the name of the chain to insert the address to
//   - insertValidators: a boolean to indicate if the address is a validator or account
//
// Returns:
//   - error: an error if the insertion fails
func (t *TimescaleDb) InsertAnAddress(address string, chainName string, insertValidators bool) error {
	column_names := []string{"address", "chain_name"}
	var table_name string
	if insertValidators {
		table_name = "gno_validators"
	} else {
		table_name = "gno_addresses"
	}
	// create interface to copy from slice to the db
	// might be an overkill for one address, but lets keep the consistency
	// and it should be technically a bit faster than using INSERT INTO...
	copy_from_slice := pgx.CopyFromSlice(1, func(i int) ([]any, error) {
		return []any{address, chainName}, nil
	})
	_, err := t.pool.CopyFrom(context.Background(), pgx.Identifier{table_name}, column_names, copy_from_slice)
	return err
}
