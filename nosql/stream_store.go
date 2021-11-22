package nosql

type streamStore struct {
	store  *Store
	stream Stream
}
