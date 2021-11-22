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
	errDocID  DocID
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
		docs:      nil,
		docsBind:  false,
		index:     nil,
		indexBind: false,
		docID:     0,
		doc:       "",
		docTyped:  nil,
		prev:      "",
		prevTyped: nil,
		buffer:    make([]byte, 8192),
		i64:       nil,
		f64:       nil,
		str:       nil,
	}
}

func (tx *Tx) Docs() *mdbx.Cursor {
	if !tx.docsBind {
		if tx.docs == nil {
			var err mdbx.Error
			tx.docs, err = tx.Tx.OpenCursor(tx.store.documentsDBI)
			if err != mdbx.ErrSuccess {
				panic(err)
			}
		} else {
			if e := tx.docs.Renew(tx.Tx); e != mdbx.ErrSuccess {
				panic(e)
			}
		}
		tx.docsBind = true
	}
	return tx.docs
}

func (tx *Tx) Index() *mdbx.Cursor {
	if !tx.indexBind {
		if tx.index == nil {
			var err mdbx.Error
			tx.index, err = tx.Tx.OpenCursor(tx.store.indexDBI)
			if err != mdbx.ErrSuccess {
				panic(err)
			}
		} else {
			if e := tx.index.Renew(tx.Tx); e != mdbx.ErrSuccess {
				panic(e)
			}
		}
		tx.indexBind = true
	}
	return tx.docs
}

func (tx *Tx) Reset(txn *mdbx.Tx) {
	tx.Tx = txn
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
	if tx.Tx != nil {

	}
}
