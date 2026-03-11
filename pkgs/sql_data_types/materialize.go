package sql_data_types

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"time"

	dbinit "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/db_init"
)

type TxCount struct {
	TimeBucket time.Time `mt:"time_bucket" fn:"time_bucket('1 hour', timestamp)" gb:"0"`
	ChainName  string    `mt:"chain_name" gb:"1"`
	Count      int64     `mt:"transaction_count" fn:"count(*)"`
}

func (tc TxCount) TableName() string {
	return "tx_count"
}

func (tc TxCount) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(tc, tc.TableName())
}

func (tc TxCount) TableColumns() []string {
	return aggColumns(tc)
}

func (tc TxCount) TableFunctions() []string {
	return aggFunctions(tc)
}

func (tc TxCount) GroupBy() []string {
	return aggGroupBy(tc)
}

// Timescaledb continuos aggregation requires source table to be specified
// this is the source table for the continuous aggregation
func (tc TxCount) FromTable() string {
	return "transaction_general"
}

// Specification for timescaledb continuous aggregation policy
//
// Usage:
// This is used to build the SQL for the aggregation policy.
//
// Returns:
//   - tableName: the name of the table to aggregate
//   - startOffset: the start offset for the aggregation
//   - endOffset: the end offset for the aggregation
//   - interval: the interval for the aggregation
func (tc TxCount) AggregatePolicy(
	startOffset *time.Duration,
	endOffset *time.Duration,
	interval *time.Duration,
) (string, string, string, string) {
	if startOffset == nil {
		d := 24 * time.Hour * 7
		startOffset = &d
	}
	if endOffset == nil {
		d := 15 * time.Second
		endOffset = &d
	}
	if interval == nil {
		d := 15 * time.Second
		interval = &d
	}
	formattedStartOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(startOffset.Seconds()), 10))
	formattedEndOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(endOffset.Seconds()), 10))
	formattedInterval := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(interval.Seconds()), 10))
	return tc.TableName(), formattedStartOffset, formattedEndOffset, formattedInterval
}

type FeeVolume struct {
	TimeBucket time.Time `mt:"time_bucket" fn:"time_bucket('1 hour', timestamp)" gb:"0"`
	ChainName  string    `mt:"chain_name" gb:"2"`
	FeeDenom   string    `mt:"denom" fn:"(fee).denom" gb:"1"`
	FeeVolume  int64     `mt:"volume" fn:"sum((fee).amount)"`
}

func (dfv FeeVolume) TableName() string {
	return "fee_volume"
}

func (dfv FeeVolume) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(dfv, dfv.TableName())
}

func (dfv FeeVolume) TableColumns() []string {
	return aggColumns(dfv)
}

func (dfv FeeVolume) TableFunctions() []string {
	return aggFunctions(dfv)
}

func (dfv FeeVolume) GroupBy() []string {
	return aggGroupBy(dfv)
}

func (dfv FeeVolume) FromTable() string {
	return "transaction_general"
}

func (dfv FeeVolume) AggregatePolicy(
	startOffset *time.Duration,
	endOffset *time.Duration,
	interval *time.Duration,
) (string, string, string, string) {
	if startOffset == nil {
		d := 24 * time.Hour * 7
		startOffset = &d
	}
	if endOffset == nil {
		d := 15 * time.Second
		endOffset = &d
	}
	if interval == nil {
		d := 15 * time.Second
		interval = &d
	}

	formattedStartOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(startOffset.Seconds()), 10))
	formattedEndOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(endOffset.Seconds()), 10))
	formattedInterval := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(interval.Seconds()), 10))

	return dfv.TableName(), formattedStartOffset, formattedEndOffset, formattedInterval
}

type DailyActiveAccounts struct {
	TimeBucket time.Time `mt:"time_bucket" fn:"time_bucket('1 day', timestamp)" gb:"0"`
	ChainName  string    `mt:"chain_name" gb:"1"`
	AccCount   int64     `mt:"active_account_count" fn:"count(DISTINCT address)"`
}

func (dac DailyActiveAccounts) TableName() string {
	return "daily_active_accounts"
}

func (dac DailyActiveAccounts) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(dac, dac.TableName())
}

func (dac DailyActiveAccounts) TableColumns() []string {
	return aggColumns(dac)
}

func (dac DailyActiveAccounts) TableFunctions() []string {
	return aggFunctions(dac)
}

func (dac DailyActiveAccounts) GroupBy() []string {
	return aggGroupBy(dac)
}

func (dac DailyActiveAccounts) FromTable() string {
	return "address_tx"
}

func (dac DailyActiveAccounts) AggregatePolicy(
	startOffset *time.Duration,
	endOffset *time.Duration,
	interval *time.Duration,
) (string, string, string, string) {
	if startOffset == nil {
		d := 24 * time.Hour * 180
		startOffset = &d
	}
	if endOffset == nil {
		d := time.Hour
		endOffset = &d
	}
	if interval == nil {
		d := 30 * time.Minute
		interval = &d
	}
	formattedStartOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(startOffset.Seconds()), 10))
	formattedEndOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(endOffset.Seconds()), 10))
	formattedInterval := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(interval.Seconds()), 10))
	return dac.TableName(), formattedStartOffset, formattedEndOffset, formattedInterval
}

type TransactionCount struct {
	TimeBucket time.Time `mt:"time_bucket" fn:"time_bucket('1 hour', timestamp)" gb:"0"`
	ChainName  string    `mt:"chain_name" gb:"1"`
	Count      int64     `mt:"transaction_count" fn:"count(*)"`
}

func (ttc TransactionCount) TableName() string {
	return "transaction_count"
}

func (ttc TransactionCount) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(ttc, ttc.TableName())
}

func (ttc TransactionCount) TableColumns() []string {
	return aggColumns(ttc)
}

func (ttc TransactionCount) TableFunctions() []string {
	return aggFunctions(ttc)
}

func (ttc TransactionCount) GroupBy() []string {
	return aggGroupBy(ttc)
}

func (ttc TransactionCount) FromTable() string {
	return "transaction_general"
}

func (ttc TransactionCount) AggregatePolicy(
	startOffset *time.Duration,
	endOffset *time.Duration,
	interval *time.Duration,
) (string, string, string, string) {
	if startOffset == nil {
		d := 24 * time.Hour * 7
		startOffset = &d
	}
	if endOffset == nil {
		d := 15 * time.Second
		endOffset = &d
	}
	if interval == nil {
		d := 15 * time.Second
		interval = &d
	}
	formattedStartOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(startOffset.Seconds()), 10))
	formattedEndOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(endOffset.Seconds()), 10))
	formattedInterval := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(interval.Seconds()), 10))
	return ttc.TableName(), formattedStartOffset, formattedEndOffset, formattedInterval
}

type ValidatorDailySigning struct {
	TimeBucket  time.Time `mt:"time_bucket" fn:"time_bucket('1 day', vbs.timestamp)" gb:"0"`
	ChainName   string    `mt:"chain_name" gb:"1"`
	ValidatorId int32     `mt:"validator_id" gb:"2"`
	BlockSigned int64     `mt:"blocks_signed" fn:"count(*)"`
}

func (vds ValidatorDailySigning) TableName() string {
	return "validator_daily_signing"
}

func (vds ValidatorDailySigning) FromTable() string {
	return "validator_block_signing"
}

func (vds ValidatorDailySigning) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(vds, vds.TableName())
}

func (vds ValidatorDailySigning) TableColumns() []string {
	return aggColumns(vds)
}

func (vds ValidatorDailySigning) TableFunctions() []string {
	return aggFunctions(vds)
}

func (vds ValidatorDailySigning) GroupBy() []string {
	return aggGroupBy(vds)
}

func (vds ValidatorDailySigning) FromTableAlias() string { return "vbs" }
func (vds ValidatorDailySigning) LateralJoins() []dbinit.LateralJoinDef {
	return []dbinit.LateralJoinDef{
		{
			Kind:       "CROSS JOIN LATERAL",
			Expression: "unnest(vbs.signed_vals)",
			Alias:      "v_id",
			Columns:    []string{"validator_id"},
		},
	}
}

func (vds ValidatorDailySigning) AggregatePolicy(
	startOffset *time.Duration,
	endOffset *time.Duration,
	interval *time.Duration,
) (string, string, string, string) {
	if startOffset == nil {
		d := 24 * time.Hour * 14
		startOffset = &d
	}
	if endOffset == nil {
		d := 15 * time.Second
		endOffset = &d
	}
	if interval == nil {
		d := 15 * time.Second
		interval = &d
	}
	formattedStartOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(startOffset.Seconds()), 10))
	formattedEndOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(endOffset.Seconds()), 10))
	formattedInterval := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(interval.Seconds()), 10))
	return vds.TableName(), formattedStartOffset, formattedEndOffset, formattedInterval
}

type DailyBlockCount struct {
	TimeBucket time.Time `mt:"time_bucket" fn:"time_bucket('1 day', timestamp)" gb:"0"`
	ChainName  string    `mt:"chain_name" gb:"1"`
	BlockCount int64     `mt:"block_count" fn:"count(*)"`
}

func (dbc DailyBlockCount) TableName() string {
	return "daily_block_count"
}

func (dbc DailyBlockCount) GetTableInfo() (*dbinit.TableInfo, error) {
	return dbinit.GetTableInfo(dbc, dbc.TableName())
}

func (dbc DailyBlockCount) TableColumns() []string {
	return aggColumns(dbc)
}

func (dbc DailyBlockCount) TableFunctions() []string {
	return aggFunctions(dbc)
}

func (dbc DailyBlockCount) GroupBy() []string {
	return aggGroupBy(dbc)
}

func (dbc DailyBlockCount) FromTable() string {
	return "blocks"
}

func (dbc DailyBlockCount) AggregatePolicy(
	startOffset *time.Duration,
	endOffset *time.Duration,
	interval *time.Duration,
) (string, string, string, string) {
	if startOffset == nil {
		d := 24 * time.Hour * 14
		startOffset = &d
	}
	if endOffset == nil {
		d := 15 * time.Second
		endOffset = &d
	}
	if interval == nil {
		d := 15 * time.Second
		interval = &d
	}
	formattedStartOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(startOffset.Seconds()), 10))
	formattedEndOffset := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(endOffset.Seconds()), 10))
	formattedInterval := fmt.Sprintf("%s seconds", strconv.FormatInt(int64(interval.Seconds()), 10))
	return dbc.TableName(), formattedStartOffset, formattedEndOffset, formattedInterval
}

func aggColumns(v any) []string {
	t := reflect.TypeOf(v)
	cols := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		cols = append(cols, t.Field(i).Tag.Get("mt"))
	}
	return cols
}
func aggFunctions(v any) []string {
	t := reflect.TypeOf(v)
	fns := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		fn := t.Field(i).Tag.Get("fn")
		if fn == "" {
			fn = "noop"
		}
		fns = append(fns, fn)
	}
	return fns
}

func aggGroupBy(v any) []string {
	t := reflect.TypeOf(v)
	gb := make([]string, t.NumField())
	for i := range t.NumField() {
		f := t.Field(i)
		name := f.Tag.Get("mt")
		idx := f.Tag.Get("gb")
		if idx == "" {
			continue
		}
		idxInt, _ := strconv.Atoi(idx)
		gb = slices.Insert(gb, idxInt, name)
	}
	return gb
}
