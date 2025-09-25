package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// Experiment 8
// Trying to insert the data using the pgx package
// using the COPY FROM method
// to use any custom type I need to reguster it first
// then I can copy into it,
// for numeric use directly the pgtype.Numeric type

type Balance struct {
	Currency string
	Amount   pgtype.Numeric
}

type Account struct {
	ID       int       `db:"id"`
	Balances []Balance `db:"balances"`
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
		dataTypeNames := []string{"balance", "_balance"}

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

	// Create account with multiple balances
	accounts := []Account{
		{ID: 1,
			Balances: []Balance{
				{Currency: "USD", Amount: numericFromDecimal(decimal.NewFromFloat(100.50))},
				{Currency: "EUR", Amount: numericFromDecimal(decimal.NewFromFloat(85.25))},
				{Currency: "GBP", Amount: numericFromDecimal(decimal.NewFromFloat(75.00))},
			},
		},
		{ID: 2,
			Balances: []Balance{
				{Currency: "USD", Amount: numericFromDecimal(decimal.NewFromFloat(83.50))},
				{Currency: "EUR", Amount: numericFromDecimal(decimal.NewFromFloat(50.25))},
				{Currency: "GBP", Amount: numericFromDecimal(decimal.NewFromFloat(25.00))},
			},
		},
	}

	// Copy from slice
	pgxSlice := pgx.CopyFromSlice(len(accounts), func(i int) ([]any, error) {
		return []any{accounts[i].ID, accounts[i].Balances}, nil
	})
	_, err = pool.CopyFrom(ctx, pgx.Identifier{"account"}, []string{"id", "balances"}, pgxSlice)
	if err != nil {
		fmt.Println(err)
	}

	// Query back
	var retrieved Account
	err = pool.QueryRow(ctx, `
        SELECT id, balances 
        FROM account WHERE id = $1`, 1).
		Scan(&retrieved.ID, &retrieved.Balances)
	if err != nil {
		fmt.Printf("Error scanning: %v\n", err)
	}
	fmt.Printf("Account ID: %d\n", retrieved.ID)
	fmt.Printf("Balances: %v\n", retrieved.Balances)

	fmt.Println("All Balances:")
	for _, bal := range retrieved.Balances {
		fmt.Printf("  %s %v, %s\n", bal.Amount.Int.String(), bal.Amount.Exp, bal.Currency)
	}
}

func numericFromDecimal(d decimal.Decimal) pgtype.Numeric {
	if d.Exponent() == 0 {
		return pgtype.Numeric{Int: d.Coefficient(), Exp: 0, Valid: true}
	}
	n := pgtype.Numeric{Int: d.Coefficient(), Exp: d.Exponent(), Valid: true}
	return n
}
