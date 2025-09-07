package sql_data_types

import dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/db_init"

// Fee is a postgres type that is used to store the fee of a transaction
//
// It is a custom type that is used to store the fee of a transaction
// Stores:
// - Amount (uint64)
// - Denom (string)
// PRIMARY KEY (amount, denom)
type Amount struct {
	Amount uint64 `db:"amount" dbtype:"NUMERIC"`
	Denom  string `db:"denom" dbtype:"TEXT"`
}

// TypeName returns the name of the type for the Fee struct
func (f Amount) TypeName() string {
	return "amount"
}

// GetSpecialTypeInfo returns the special type info for the Fee struct
func (f Amount) GetSpecialTypeInfo() (*dbinit.SpecialType, error) {
	return dbinit.GetSpecialTypeInfo(f, f.TypeName())
}

// Attribute is a postgres type that is used to store the attribute of an event
//
// It is a custom type that is used to store the attribute of an event
// Stores:
// - Key (string)
// - Value (string)
type Attribute struct {
	Key   string `db:"key" dbtype:"TEXT"`
	Value string `db:"value" dbtype:"TEXT"`
}

// TypeName returns the name of the type for the Attribute struct
func (a Attribute) TypeName() string {
	return "attribute"
}

// GetSpecialTypeInfo returns the special type info for the Attribute struct
func (a Attribute) GetSpecialTypeInfo() (*dbinit.SpecialType, error) {
	return dbinit.GetSpecialTypeInfo(a, a.TypeName())
}

// Event is a postgres type that is used to store the event of a transaction
//
// Usage:
//
//	This is used to store transaction events in "native format"
//
// It is a custom type that is used to store the event of a transaction
// Stores:
// - AtType (string)
// - Type (string)
// - Attributes (Attribute[])
// - PkgPath (string)
type Event struct {
	AtType     string      `db:"at_type" dbtype:"TEXT"`
	Type       string      `db:"type" dbtype:"TEXT"`
	Attributes []Attribute `db:"attributes" dbtype:"attribute[]"`
	PkgPath    string      `db:"pkg_path" dbtype:"TEXT"`
}

// TypeName returns the name of the type for the Event struct
func (e Event) TypeName() string {
	return "event"
}

// GetSpecialTypeInfo returns the special type info for the Event struct
func (e Event) GetSpecialTypeInfo() (*dbinit.SpecialType, error) {
	return dbinit.GetSpecialTypeInfo(e, e.TypeName())
}

type DataTypes interface {
	TableName(valTable bool) string
	TypeName() string
	GetTableInfo() (*dbinit.TableInfo, error)
}

// DBSpecialType is an interface for structs that represent custom database types
type DBSpecialType interface {
	GetSpecialTypeInfo() (*dbinit.SpecialType, error)
	TypeName() string
}
