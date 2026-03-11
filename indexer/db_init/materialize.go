package dbinit

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// LateralJoinDef describes a LATERAL JOIN clause used in a continuous aggregate query.
type LateralJoinDef struct {
	Kind       string   // e.g. "CROSS JOIN LATERAL"
	Expression string   // e.g. "unnest(vbs.signed_vals)"
	Alias      string   // e.g. "v_id"
	Columns    []string // e.g. ["validator_id"]
}

// ContinuousAggregateDefinition is the interface that materialize structs must implement
// to be used with the DBInitializer continuous aggregate helpers.
type ContinuousAggregateDefinition interface {
	// TableName returns the name of the materialized view.
	TableName() string
	// FromTable returns the source hypertable for the aggregation.
	FromTable() string
	// TableColumns returns the ordered list of output column aliases (mt tag values).
	TableColumns() []string
	// TableFunctions returns the SQL expression for each column ("noop" means passthrough).
	TableFunctions() []string
	// GroupBy returns the ordered list of GROUP BY column aliases (gb tag values).
	GroupBy() []string
	// AggregatePolicy returns the view name and formatted policy intervals.
	AggregatePolicy(startOffset, endOffset, interval *time.Duration) (string, string, string, string)
}

// LateralJoiner is an optional extension of ContinuousAggregateDefinition for views
// that require a LATERAL JOIN (e.g. unnesting arrays).
type LateralJoiner interface {
	FromTableAlias() string
	LateralJoins() []LateralJoinDef
}

// GenerateContinuousAggregateSQL builds the CREATE MATERIALIZED VIEW … WITH
// (timescaledb.continuous) SQL for the given aggregate definition.
//
// GROUP BY uses 1-based ordinal positions derived from the column order so the
// generated SQL is unambiguous regardless of whether a name shadows a built-in.
func GenerateContinuousAggregateSQL(agg ContinuousAggregateDefinition) string {
	cols := agg.TableColumns()
	fns := agg.TableFunctions()
	groupByCols := agg.GroupBy()

	// Build SELECT clause.
	selectParts := make([]string, 0, len(cols))
	for i, col := range cols {
		fn := ""
		if i < len(fns) {
			fn = fns[i]
		}
		if fn == "" || fn == "noop" {
			selectParts = append(selectParts, col)
		} else {
			selectParts = append(selectParts, fmt.Sprintf("%s AS %s", fn, col))
		}
	}

	// Build GROUP BY clause using 1-based ordinal positions.
	colOrdinals := make(map[string]int, len(cols))
	for i, c := range cols {
		colOrdinals[c] = i + 1
	}
	gbOrdinals := make([]string, 0, len(groupByCols))
	for _, gb := range groupByCols {
		if gb == "" {
			continue
		}
		if ord, ok := colOrdinals[gb]; ok {
			gbOrdinals = append(gbOrdinals, strconv.Itoa(ord))
		}
	}

	// Build FROM clause, including optional alias and LATERAL joins.
	fromClause := agg.FromTable()
	lateralParts := []string{}
	if lj, ok := agg.(LateralJoiner); ok {
		fromClause = fmt.Sprintf("%s %s", agg.FromTable(), lj.FromTableAlias())
		for _, join := range lj.LateralJoins() {
			colList := strings.Join(join.Columns, ", ")
			lateralParts = append(lateralParts, fmt.Sprintf(
				"%s %s AS %s(%s)",
				join.Kind, join.Expression, join.Alias, colList,
			))
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "CREATE MATERIALIZED VIEW %s\n", agg.TableName())
	sb.WriteString("WITH (timescaledb.continuous) AS\n")
	sb.WriteString("SELECT\n")
	sb.WriteString("    " + strings.Join(selectParts, ",\n    ") + "\n")
	fmt.Fprintf(&sb, "FROM %s\n", fromClause)
	for _, lp := range lateralParts {
		sb.WriteString(lp + "\n")
	}
	if len(gbOrdinals) > 0 {
		fmt.Fprintf(&sb, "GROUP BY %s\n", strings.Join(gbOrdinals, ", "))
	}
	sb.WriteString(";")
	return sb.String()
}

// CreateContinuousAggregate executes the CREATE MATERIALIZED VIEW statement for the
// given aggregate definition. Errors are logged but not fatal.
func (dbi *DBInitializer) CreateContinuousAggregate(agg ContinuousAggregateDefinition) error {
	sql := GenerateContinuousAggregateSQL(agg)
	_, err := dbi.pool.Exec(context.Background(), sql)
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Str("view", agg.TableName()).
			Msgf("failed to create continuous aggregate: %v", err)
		return err
	}
	return nil
}

// AlterContinuousAggregateColumnstore enables TimescaleDB columnstore compression on
// a continuous aggregate view, ordering by time_bucket DESC and segmenting by the
// provided columns.
func (dbi *DBInitializer) AlterContinuousAggregateColumnstore(viewName string, segmentByCols []string) error {
	segmentBy := strings.Join(segmentByCols, ", ")
	sql := fmt.Sprintf(
		`ALTER MATERIALIZED VIEW %s SET (
	timescaledb.enable_columnstore = TRUE,
	timescaledb.segmentby = '%s',
	timescaledb.orderby = 'time_bucket DESC'
)`,
		viewName, segmentBy,
	)
	_, err := dbi.pool.Exec(context.Background(), sql)
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Str("view", viewName).
			Msgf("failed to alter continuous aggregate columnstore: %v", err)
		return err
	}
	return nil
}

// AddContinuousAggregatePolicy registers a TimescaleDB refresh policy for the given
// continuous aggregate. The interval strings must be in the format "N seconds".
func (dbi *DBInitializer) AddContinuousAggregatePolicy(viewName, startOffset, endOffset, scheduleInterval string) error {
	sql := fmt.Sprintf(
		`SELECT add_continuous_aggregate_policy('%s',
	start_offset => INTERVAL '%s',
	end_offset   => INTERVAL '%s',
	schedule_interval => INTERVAL '%s'
)`,
		viewName, startOffset, endOffset, scheduleInterval,
	)
	_, err := dbi.pool.Exec(context.Background(), sql)
	if err != nil {
		l.Error().
			Caller().
			Stack().
			Str("view", viewName).
			Msgf("failed to add continuous aggregate policy: %v", err)
		return err
	}
	return nil
}
