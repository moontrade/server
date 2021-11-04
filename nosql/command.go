package nosql

type UpdateSchemaCommand struct {
	Collection FlatString255
	Data       []byte
}

type UpdateSchemaResult struct {
	Code int64
	Data []byte
}

type InsertCommand struct {
	Collection FlatString32
	Data       []byte
}

type UpdateCommand struct {
	Collection FlatString32
	Data       []byte
}
