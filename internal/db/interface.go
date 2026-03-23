package db

// Driver exposes the methods to interact with a database.
// Common methods are meant to perform operations in tables with
// only two columns: id and data, and both are strings.
// Blob methods are meant for tables in which the data column is
// a byte array.
// Errors are returned when the underlying driver has a problem or
// when the parameters provided are not suitable for the described
// tables.
type Driver interface {
	// Close should be called as a deferred method after driver creation
	Close()
	// Retrieve returns an empty string if the element does not exist
	Retrieve(inputType string, id string) (output string, err error)
	// RetrieveAll returns an empty string if there are no elements
	RetrieveAll(inputType string) (output map[string]string, err error)
	Create(inputType string, id string, input string) (err error)
	Update(inputType string, id string, input string) (err error)
	Remove(inputType string, id string) (err error)
	// GetContextSymbol returns the character used to separate the table name
	// from the context string
	GetContextSymbol() string

	BlobRetrieve(id string) (output *[]byte, err error)
	BlobCreate(id string, input *[]byte) (err error)
	BlobUpdate(id string, input *[]byte) (err error)
	BlobRemove(id string) (err error)
	BlobExists(id string) (output bool, err error)
}
