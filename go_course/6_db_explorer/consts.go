package main

const (
	defaultLimit  = 5
	defaultOffset = 0
)

const (
	getTablesQuery   = "SHOW TABLES"
	getPKColumnQuery = "SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_NAME='%s' AND TABLE_SCHEMA='%s' AND COLUMN_KEY='PRI'"
	getColumnsQuery  = "SELECT COLUMN_NAME, IS_NULLABLE, DATA_TYPE, COLUMN_KEY, EXTRA FROM information_schema.COLUMNS WHERE TABLE_NAME='%s' AND TABLE_SCHEMA='%s'"

	selectFromTableQuery = "SELECT * FROM %s"
	limitQuery           = " LIMIT %d"
	offsetQuery          = " OFFSET %d"

	selectByIDQuery = "SELECT * FROM %s WHERE %s = ?"
)

const (
	tableSchema        = "explorer"
	autoIncrementExtra = "auto_increment"
	pkColumnLabel      = "PRI"
)

const (
	invalidTypeErrMsg = "field %s have invalid type"
)
