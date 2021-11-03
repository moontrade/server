package logger

import "github.com/moontrade/mdbx-go"

type MDBXWriter struct {
	store *mdbx.Store
}
