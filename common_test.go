package elephant

const temporaryDB = "/tmp/foo.db"

type structCheck struct {
	Mystring string
	Myint    int
	Myint64  int64 `db:"key"`
	Mybool   bool
}

type failingStructCheck struct {
	Mystring string
	Myint    int `db:"key"`
	Myint64  int64
	Mybool   bool
}
