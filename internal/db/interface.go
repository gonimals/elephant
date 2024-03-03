package db

type Driver interface {
	Close()
	Retrieve(inputType string, key string) (output string, err error)
	RetrieveAll(inputType string) (output map[string]string, err error)
	Create(inputType string, key string, input string) (err error)
	Update(inputType string, key string, input string) (err error)
	Remove(inputType string, key string) (err error)

	BlobRetrieve(key string) (output *[]byte, err error)
	BlobCreate(key string, input *[]byte) (err error)
	BlobUpdate(key string, input *[]byte) (err error)
	BlobRemove(key string) (err error)
}
