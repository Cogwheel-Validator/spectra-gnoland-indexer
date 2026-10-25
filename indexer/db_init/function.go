package dbinit

import (
	"context"
	"fmt"
)

// FunctionDef describes a single CREATE OR REPLACE FUNCTION statement.
//
// Functions are hand-written SQL, there is nothing to derive from struct tags.
// Declare them next to the tables they work on.
type FunctionDef struct {
	// Name is the function name, used for logging.
	Name string
	// SQL is the full CREATE OR REPLACE FUNCTION statement.
	SQL string
}

// CreateFunction executes the CREATE OR REPLACE FUNCTION statement for the given FunctionDef.
func (db *DBInitializer) CreateFunction(fn FunctionDef) error {
	_, err := db.pool.Exec(context.Background(), fn.SQL)
	if err != nil {
		return fmt.Errorf("failed to create function %s: %w", fn.Name, err)
	}

	l.Info(fmt.Sprintf("Successfully created function: %s", fn.Name))
	return nil
}
