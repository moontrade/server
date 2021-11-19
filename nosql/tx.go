package nosql

import "github.com/moontrade/mdbx-go"

type Tx struct {
	Tx        *mdbx.Tx
	store     *Store
	docs      *mdbx.Cursor
	docsBind  bool
	index     *mdbx.Cursor
	indexBind bool
	docID     DocID
	doc       string
	docTyped  interface{}
	prev      string
	prevTyped interface{}
	buffer    []byte
	i64       []int64
	f64       []float64
	str       []string
}

func NewTx(s *Store) *Tx {
	return &Tx{
		store:     s,
		docs:      mdbx.NewCursor(),
		docsBind:  false,
		index:     mdbx.NewCursor(),
		indexBind: false,
		docID:     0,
		doc:       "",
		docTyped:  nil,
		prev:      "",
		prevTyped: nil,
		buffer:    nil,
		i64:       nil,
		f64:       nil,
		str:       nil,
	}
}

func (tx *Tx) Docs() *mdbx.Cursor {
	if !tx.docsBind {
		if e := tx.Tx.Bind(tx.docs, tx.store.documentsDBI); e != mdbx.ErrSuccess {
			panic(e)
		}
		tx.docsBind = true
	}
	return tx.docs
}

func (tx *Tx) Index() *mdbx.Cursor {
	if !tx.indexBind {
		if e := tx.Tx.Bind(tx.index, tx.store.indexDBI); e != mdbx.ErrSuccess {
			panic(e)
		}
		tx.indexBind = true
	}
	return tx.docs
}

func (tx *Tx) Reset() {
	tx.docsBind = false
	tx.indexBind = false
	tx.doc = ""
	tx.docTyped = nil
	tx.prev = ""
	tx.prevTyped = nil
}

func (tx *Tx) Close() {
	if tx.docs != nil {
		_ = tx.docs.Close()
		tx.docs = nil
	}
	if tx.index != nil {
		_ = tx.index.Close()
		tx.index = nil
	}
	tx.Reset()
}
