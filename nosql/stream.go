package nosql

// Stream is a special type of collection that provides real-time streaming
// similar to Kafka. Streams consist of pages which contain individual records.
// Appending new records occurs through an Appender. Appender is responsible for
// building new immutable pages and notifying listeners in real-time. Appender
// does not clear any records until the page the records belong to have been
// confirmed to be persistent in secondary storage.
//
// Secondary page storage can be any KV store. Since pages are immutable, it
// is very easy to add caching layers if needed.
//
// Pages can be located by RecordID or by Timestamp.
type Stream struct {
}

type Record struct {
	ID   DocID
	Time int64
}

type Page struct {
	ID DocID
}
