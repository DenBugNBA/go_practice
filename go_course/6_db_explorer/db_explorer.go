package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

type dbExplorer struct {
	db *sql.DB
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	explorer := dbExplorer{db: db}

	indexMux := http.NewServeMux()
	indexMux.HandleFunc("/", explorer.handleListTables())

	tableMux := http.NewServeMux()
	tableMux.HandleFunc("/{table}/", explorer.handleTable())
	tableMux.HandleFunc("/{table}/{id}", explorer.handleTableRecord())

	tableHandler := explorer.validateTableMiddleware(tableMux)

	mainMux := http.NewServeMux()
	mainMux.Handle("/", indexMux)
	mainMux.Handle("/{table}/", tableHandler)

	return mainMux, nil
}

func (e *dbExplorer) validateTableMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		table := r.PathValue("table")
		if err := e.validateTable(r.Context(), table); err != nil {
			switch {
			case errors.Is(err, errTableNotFound):
				e.writeResp(http.StatusNotFound, &errorResp{Error: "unknown table"}, w)
			default:
				e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (e *dbExplorer) validateTable(ctx context.Context, tableName string) error {
	allTables, err := e.getTables(ctx)
	if err != nil {
		return fmt.Errorf("get tables: %w", err)
	}
	if !slices.Contains(allTables, tableName) {
		return errTableNotFound
	}
	return nil
}

func (e *dbExplorer) writeResp(statusCode int, resp any, w http.ResponseWriter) {
	bytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("marshal response")
		return
	}

	w.WriteHeader(statusCode)

	_, err = w.Write(bytes)
	if err != nil {
		log.Error().Err(err).Msg("write http response")
	}
}

func (e *dbExplorer) handleListTables() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tables, err := e.getTables(r.Context())
		if err != nil {
			e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
			log.Error().Err(err).Msg("get tables")
			return
		}
		e.writeResp(http.StatusOK, &respWrapper{Resp: listTablesResp{Tables: tables}}, w)
	}
}

func (e *dbExplorer) getTables(ctx context.Context) ([]string, error) {
	rows, err := e.db.QueryContext(ctx, getTablesQuery)
	if err != nil {
		return nil, fmt.Errorf("show tables: %w", err)
	}
	defer func() { logClose(rows, "rows") }()
	var tables []string
	for rows.Next() {
		var table string
		if err = rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (e *dbExplorer) handleTable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			e.getRows(w, r)
		case http.MethodPut:
			e.createRow(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func (e *dbExplorer) getRows(w http.ResponseWriter, r *http.Request) {
	ps, err := e.parseSelectParams(r)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	records, err := e.fetchRecords(r.Context(), ps)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	e.writeResp(http.StatusOK, &respWrapper{Resp: &tableResp{Records: records}}, w)
}

func (e *dbExplorer) parseSelectParams(r *http.Request) (*selectParams, error) {
	ps := &selectParams{TableName: r.PathValue("table"), Limit: defaultLimit, Offset: defaultOffset}

	q := r.URL.Query()

	if limitStr := q.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			ps.Limit = limit
		}
	}

	if offsetStr := q.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			ps.Offset = offset
		}
	}

	return ps, nil
}

func (e *dbExplorer) fetchRecords(ctx context.Context, ps *selectParams) ([]map[string]any, error) {
	rs, err := e.db.QueryContext(ctx, e.buildFetchRecordsQuery(ps))
	if err != nil {
		return nil, fmt.Errorf("fetch records: %w", err)
	}
	defer func() { logClose(rs, "rows") }()

	tableColumns, err := e.parseTableColumns(rs)
	if err != nil {
		return nil, err
	}

	var records []map[string]any
	for rs.Next() {
		vals := e.createRowEmptyValues(tableColumns)
		if err = rs.Scan(vals...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		records = append(records, e.createRowRecord(vals, tableColumns))
	}
	return records, nil
}

func (e *dbExplorer) buildFetchRecordsQuery(ps *selectParams) string {
	sb := new(strings.Builder)
	sb.WriteString(fmt.Sprintf(selectFromTableQuery, ps.TableName))
	if ps.Limit > 0 {
		sb.WriteString(fmt.Sprintf(limitQuery, ps.Limit))
	}
	if ps.Offset > 0 {
		sb.WriteString(fmt.Sprintf(offsetQuery, ps.Offset))
	}
	return sb.String()
}

func (e *dbExplorer) parseTableColumns(rs *sql.Rows) ([]*tableColumn, error) {
	tps, err := rs.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("fetch column types: %w", err)
	}
	tableColumns := make([]*tableColumn, 0, len(tps))
	for _, t := range tps {
		tableColumns = append(tableColumns, &tableColumn{Name: t.Name(), Type: t.ScanType()})
	}
	return tableColumns, nil
}

func (e *dbExplorer) createRowEmptyValues(tableColumns []*tableColumn) []any {
	res := make([]any, 0, len(tableColumns))
	for _, c := range tableColumns {
		res = append(res, createValueByType(c.Type))
	}
	return res
}

func (e *dbExplorer) createRowRecord(vals []any, tableColumns []*tableColumn) map[string]any {
	res := make(map[string]any, len(tableColumns))
	for i, col := range tableColumns {
		res[col.Name] = formatValue(vals[i])
	}
	return res
}

func (e *dbExplorer) createRow(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")

	columns, err := e.fetchExtTableColumns(r.Context(), table)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	fields, err := e.parseReqFields(r)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	pkColumn, err := e.getPKColumn(columns)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	q, args := e.buildInsertQuery(table, pkColumn, columns, fields)

	var id any
	err = e.db.QueryRowContext(r.Context(), q, args...).Scan(&id)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	e.writeResp(http.StatusOK, &respWrapper{Resp: map[string]any{pkColumn: id}}, w)
}

func (e *dbExplorer) fetchExtTableColumns(ctx context.Context, table string) ([]*extendedTableColumn, error) {
	rs, err := e.db.QueryContext(ctx, fmt.Sprintf(getColumnsQuery, table, tableSchema))
	if err != nil {
		return nil, fmt.Errorf("fetch columns: %w", err)
	}
	defer func() { logClose(rs, "rows") }()

	var columns []*extendedTableColumn
	for rs.Next() {
		c := new(extendedTableColumn)
		if err = rs.Scan(&c.Name, &c.IsNullable, &c.Type, &c.ColumnKey, &c.Extra); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		columns = append(columns, c)
	}
	return columns, nil
}

func (e *dbExplorer) parseReqFields(r *http.Request) (map[string]any, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read req body: %w", err)
	}
	defer func() { logClose(r.Body, "req body") }()

	fields := make(map[string]any)
	err = json.Unmarshal(body, &fields)
	if err != nil {
		return nil, fmt.Errorf("unmarshal req body: %w", err)
	}

	return fields, nil
}

func (e *dbExplorer) getPKColumn(columns []*extendedTableColumn) (string, error) {
	for _, col := range columns {
		if col.ColumnKey == pkColumnLabel {
			return col.Name, nil
		}
	}
	return "", fmt.Errorf("no column with label '%s'", pkColumnLabel)
}

func (e *dbExplorer) buildInsertQuery(
	table string,
	pkColumn string,
	columns []*extendedTableColumn,
	fields map[string]any,
) (string, []any) {
	sb := new(strings.Builder)
	sb.WriteString("INSERT INTO ")
	sb.WriteString(table)

	insertFields := make([]string, 0)
	insertValues := make([]any, 0)

	for _, c := range columns {
		// AUTO_INCREMENT primary key игнорируется при вставке
		if c.ColumnKey == pkColumnLabel && strings.Contains(c.Extra, autoIncrementExtra) {
			continue
		}

		if v := fields[c.Name]; v != nil {
			insertFields = append(insertFields, c.Name)
			insertValues = append(insertValues, v)
		} else if c.IsNullable == "NO" {
			insertFields = append(insertFields, c.Name)
			insertValues = append(insertValues, createDefaultValByType(c.Type))
		}
	}
	insertFieldsCnt := len(insertFields)

	sb.WriteString(" (")
	for i, field := range insertFields {
		sb.WriteString(field)
		if i < insertFieldsCnt-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")

	sb.WriteString(" VALUES (")
	for i := range insertValues {
		sb.WriteString("?")
		if i < insertFieldsCnt-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")

	sb.WriteString(" RETURNING ")
	sb.WriteString(pkColumn)

	return sb.String(), insertValues
}

func (e *dbExplorer) handleTableRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			e.getRecord(w, r)
		case http.MethodPost:
			e.updateRecord(w, r)
		case http.MethodDelete:
			e.deleteRecord(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func (e *dbExplorer) getRecord(w http.ResponseWriter, r *http.Request) {
	ps := &tableRecordParams{Name: r.PathValue("table"), Id: r.PathValue("id")}
	rec, err := e.fetchRecord(r.Context(), ps)
	if errors.Is(err, errRecNotFound) {
		e.writeResp(http.StatusNotFound, &errorResp{Error: err.Error()}, w)
		return
	} else if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	e.writeResp(http.StatusOK, &respWrapper{Resp: &tableRecordResp{Record: rec}}, w)
}

func (e *dbExplorer) fetchRecord(ctx context.Context, ps *tableRecordParams) (map[string]any, error) {
	pkColumn, err := e.fetchPkColumn(ctx, ps)
	if err != nil {
		return nil, fmt.Errorf("get pk column: %w", err)
	}

	rs, err := e.db.QueryContext(ctx, fmt.Sprintf(selectByIDQuery, ps.Name, pkColumn), ps.Id)
	if err != nil {
		return nil, fmt.Errorf("select record: %w", err)
	}
	defer func() { logClose(rs, "rows") }()

	tableColumns, err := e.parseTableColumns(rs)
	if err != nil {
		return nil, err
	}

	hasRec := rs.Next()
	if !hasRec {
		return nil, errRecNotFound
	}

	vals := e.createRowEmptyValues(tableColumns)
	if err = rs.Scan(vals...); err != nil {
		return nil, fmt.Errorf("scan row: %w", err)
	}
	return e.createRowRecord(vals, tableColumns), nil
}

func (e *dbExplorer) fetchPkColumn(ctx context.Context, ps *tableRecordParams) (string, error) {
	var pkColumn string
	err := e.db.QueryRowContext(ctx, fmt.Sprintf(getPKColumnQuery, ps.Name, tableSchema)).Scan(&pkColumn)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("table %s returned no PK", ps.Name)
	} else if err != nil {
		return "", fmt.Errorf("show PK: %w", err)
	}
	return pkColumn, nil
}

func (e *dbExplorer) updateRecord(w http.ResponseWriter, r *http.Request) {
	ps := &tableRecordParams{Name: r.PathValue("table"), Id: r.PathValue("id")}

	columns, err := e.fetchExtTableColumns(r.Context(), ps.Name)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	fields, err := e.parseReqFields(r)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	pkColumn, err := e.getPKColumn(columns)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	err = e.validateRecordUpdate(columns, fields)
	if err != nil {
		e.writeResp(http.StatusBadRequest, &errorResp{Error: err.Error()}, w)
		return
	}

	q, args := e.buildUpdateQuery(ps, pkColumn, columns, fields)

	res, err := e.db.ExecContext(r.Context(), q, args...)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	e.writeResp(http.StatusOK, &respWrapper{Resp: &updateTableRecordResp{Updated: rowsAffected}}, w)
}

func (e *dbExplorer) validateRecordUpdate(columns []*extendedTableColumn, fields map[string]any) error {
	for _, c := range columns {
		if v, ok := fields[c.Name]; ok {
			if isNil(v) && c.IsNullable == "NO" {
				return fmt.Errorf(invalidTypeErrMsg, c.Name)
			}
			if c.ColumnKey == pkColumnLabel {
				return fmt.Errorf(invalidTypeErrMsg, c.Name)
			}
			if err := validateDataType(v, c); err != nil {
				return fmt.Errorf(invalidTypeErrMsg, c.Name)
			}
		}
	}
	return nil
}

func (e *dbExplorer) buildUpdateQuery(
	ps *tableRecordParams,
	pkColumn string,
	columns []*extendedTableColumn,
	fields map[string]any,
) (string, []any) {
	sb := new(strings.Builder)
	sb.WriteString("UPDATE ")
	sb.WriteString(ps.Name)
	sb.WriteString(" SET ")

	updateFields := make([]string, 0)
	args := make([]any, 0)

	for _, c := range columns {
		if v, ok := fields[c.Name]; ok {
			updateFields = append(updateFields, c.Name)
			args = append(args, v)
		}
	}

	updateFieldsCnt := len(updateFields)

	for i, field := range updateFields {
		sb.WriteString(field)
		sb.WriteString(" = ?")
		if i < updateFieldsCnt-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(" WHERE ")
	sb.WriteString(pkColumn)
	sb.WriteString(" = ?")
	args = append(args, ps.Id)

	return sb.String(), args
}

func (e *dbExplorer) deleteRecord(w http.ResponseWriter, r *http.Request) {
	ps := &tableRecordParams{Name: r.PathValue("table"), Id: r.PathValue("id")}

	columns, err := e.fetchExtTableColumns(r.Context(), ps.Name)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	pkColumn, err := e.getPKColumn(columns)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	q, args := e.buildDeleteQuery(ps, pkColumn)

	res, err := e.db.ExecContext(r.Context(), q, args...)
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		e.writeResp(http.StatusInternalServerError, &errorResp{Error: err.Error()}, w)
		return
	}

	e.writeResp(http.StatusOK, &respWrapper{Resp: &deleteTableRecordResp{Deleted: rowsAffected}}, w)
}

func (e *dbExplorer) buildDeleteQuery(ps *tableRecordParams, pkColumn string) (string, []any) {
	sb := new(strings.Builder)
	sb.WriteString("DELETE FROM ")
	sb.WriteString(ps.Name)
	sb.WriteString(" WHERE ")
	sb.WriteString(pkColumn)
	sb.WriteString(" = ?")
	return sb.String(), []any{ps.Id}
}
