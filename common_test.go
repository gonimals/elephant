package elephant

const temporaryDB = "/tmp/foo.db"

type structCheck struct {
	Mystring string `db:"key"`
	Myint    int
	Myint64  int64
	Mybool   bool
}

type failingStructCheck struct {
	Mystring string
	Myint    int `db:"key"`
	Myint64  int64
	Mybool   bool
}

type fileStructCheck struct {
	Filename string `db:"key"`
	Contents []byte
}

/*
type stringStructCheck struct {
	Mystring string `db:"key"`
	Mydate   int
}
*/
