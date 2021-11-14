package nosql

import "github.com/moontrade/mdbx-go"

type Tx struct {
	Tx           *mdbx.Tx
	buffer       []byte
	i64Buffer    []int64
	f64Buffer    []float64
	stringBuffer []string
}
