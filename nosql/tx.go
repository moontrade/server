package nosql

import "github.com/moontrade/mdbx-go"

type Tx struct {
	tx *mdbx.Tx
}
