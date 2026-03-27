package main

type respWrapper struct {
	Resp any `json:"response"`
}

type listTablesResp struct {
	Tables []string `json:"tables"`
}

type errorResp struct {
	Error string `json:"error"`
}

type tableResp struct {
	Records []map[string]any `json:"records"`
}

type tableRecordResp struct {
	Record map[string]any `json:"record"`
}

type updateTableRecordResp struct {
	Updated int64 `json:"updated"`
}

type deleteTableRecordResp struct {
	Deleted int64 `json:"deleted"`
}
