package main

import "reflect"

type tableColumn struct {
	Name string
	Type reflect.Type
}

type extendedTableColumn struct {
	Name       string
	IsNullable string
	Type       string
	ColumnKey  string
	Extra      string
}

type selectParams struct {
	TableName string
	Limit     int
	Offset    int
}

type tableRecordParams struct {
	Name string
	Id   string
}
