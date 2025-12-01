package main

import (
	"database/sql"
	"fmt"
	"io"
	"reflect"

	"github.com/rs/zerolog/log"
)

func logClose(closer io.Closer, obj string) {
	if err := closer.Close(); err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("close %s", obj))
	}
}

func createValueByType(dbType reflect.Type) any {
	return reflect.New(dbType).Interface()
}

func formatValue(v any) any {
	switch v.(type) {
	case *sql.NullString:
		nullStr := v.(*sql.NullString)
		if nullStr.Valid {
			return nullStr.String
		}
		return nil
	default:
		return v
	}
}

func isNil(i any) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice, reflect.Func:
		return reflect.ValueOf(i).IsNil()
	default:
		return false
	}
}

func createDefaultValByType(dataType string) any {
	switch dataType {
	case "varchar", "text":
		return ""
	case "int":
		return 0
	}
	return ""
}

func validateDataType(v any, c *extendedTableColumn) error {
	if v == nil {
		return nil
	}
	switch c.Type {
	case "varchar", "text":
		if reflect.TypeOf(v).Kind() != reflect.String {
			return fmt.Errorf("expected string type, got %s for column %s", c.Type, c.Name)
		}
	}
	return nil
}
