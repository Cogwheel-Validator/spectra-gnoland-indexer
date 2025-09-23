package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Experiment 9
// Trying to insert the data

type Attribute struct {
	Key   string
	Value string
}

type Event struct {
	AtType     string      `db:"at_type" dbtype:"TEXT"`
	Type       string      `db:"type" dbtype:"TEXT"`
	Attributes []Attribute `db:"attributes" dbtype:"attribute[]"`
	PkgPath    string      `db:"pkg_path" dbtype:"TEXT"`
}

// table for testing
/*
CREATE TYPE attribute AS (
    key TEXT,
    value TEXT
);

CREATE TYPE event AS (
	at_type TEXT,
	type TEXT,
	attributes attribute[],
	pkg_path TEXT
);

CREATE TABLE account (
    id INTEGER,
    events event[]
);
*/
type Account struct {
	ID     int     `db:"id"`
	Events []Event `db:"events"`
}

func setupConnection() (*pgxpool.Pool, error) {
	host := "localhost"
	port := 6543
	user := "postgres"
	password := "12345678"
	dbname := "gnoland"
	sslmode := "disable"

	config, err := pgxpool.ParseConfig(
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode),
	)
	if err != nil {
		return nil, err
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		dataTypeNames := []string{"attribute", "_attribute", "event", "_event"}

		for _, typeName := range dataTypeNames {
			dataType, err := conn.LoadType(ctx, typeName)
			if err != nil {
				return err
			}
			conn.TypeMap().RegisterType(dataType)
		}

		return nil
	}

	return pgxpool.NewWithConfig(context.Background(), config)
}

func main() {
	pool, err := setupConnection()
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Create a synthetic data for testing
	accounts := []Account{
		{
			ID: 1,
			Events: []Event{
				{
					AtType:  "tm.GnoEvent",
					Type:    "send",
					PkgPath: "gno.land/r/bank",
					Attributes: []Attribute{
						{Key: "sender", Value: "gno1..."},
						{Key: "recipient", Value: "gno2..."},
						{Key: "amount", Value: "1000ugnot"},
					},
				},
				{
					AtType: "tm.GnoEvent",
					Type:   "receive",
					// Let's try without the pkgpath
					Attributes: []Attribute{
						{Key: "sender", Value: "gno2..."},
						{Key: "recipient", Value: "gno1..."},
						{Key: "amount", Value: "500ugnot"},
					},
				},
			},
		},
		{
			ID:     2,
			Events: []Event{}, // empty events
		},
	}

	// Copy from slice
	pgxSlice := pgx.CopyFromSlice(len(accounts), func(i int) ([]any, error) {
		return []any{accounts[i].ID, accounts[i].Events}, nil
	})

	_, err = pool.CopyFrom(ctx, pgx.Identifier{"account"}, []string{"id", "events"}, pgxSlice)
	if err != nil {
		panic(fmt.Errorf("failed to copy data: %w", err))
	}

	fmt.Println("Successfully inserted account data.")

	// Query back to verify
	var retrieved Account
	err = pool.QueryRow(ctx, `SELECT id, events FROM account WHERE id = $1`, 1).Scan(&retrieved.ID, &retrieved.Events)
	if err != nil {
		panic(fmt.Errorf("failed to query data: %w", err))
	}

	fmt.Printf("Retrieved Account ID: %d\n", retrieved.ID)
	for i, event := range retrieved.Events {
		fmt.Printf("  Event %d: Type=%s, PkgPath=%s\n", i+1, event.Type, event.PkgPath)
		for _, attr := range event.Attributes {
			fmt.Printf("    Attribute: %s = %s\n", attr.Key, attr.Value)
		}
	}
}
