_Elephants never forget_

[![Go Reference](https://pkg.go.dev/badge/github.com/gonimals/elephant.svg)](https://pkg.go.dev/github.com/gonimals/elephant)

# Schema
The Elephant library will work only with structures which meet the following criteria:

- It has a parameter of type string which has the tag `db:"id"` and is unique (will be used as primary key)
- The structure can be marshalled to JSON
- The struct name meets the following regular expression: `[0-9A-Za-z_]{1,40}`
- All attributes to be saved must be public (first letter of the variable name must be uppercase)

This library will store every instance inside a table with the name of the structure. Each table will have two columns: the id (string) column and the value, which will be a JSON with, at most, 64 Kilobytes (defined by MaxStructLength)

Supported URIs, right now, follow this criteria:

- `sqlite3:path/to/file.db` (if the file doesn't exist, it will be created)
- `mysql:user:password@tcp(hostname:port)/database`

# Example usage
```golang
err := elephant.Initialize("sqlite3:example.db")
if err != nil {
    log.Fatalln("Initialization failed:", err)
}
defer elephant.Close()

element := struct{
    id string `db:"id"`
    dataInt int
    dataString string
} {
    dataInt: 4,
    dataString: "test",
}

elephant.Store(element)
```