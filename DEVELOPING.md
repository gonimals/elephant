# Compiling and testing
```
go get -u ./...
go test -cover ./...
go mod tidy

sqlite3 /tmp/foo.db
.tables
select * from structCheck;
```

## MySQL
To perform MySQL tests, create a docker container with the following configuration:
```bash
docker run -d -p33060:3306 --name elephant-testing \
-e MARIADB_ALLOW_EMPTY_ROOT_PASSWORD=true \
-e MARIADB_DATABASE=elephant \
mariadb:lts
```

Run it everytime the checks must be run:

```bash
docker start elephant-testing
```

You can check the access to the instance with:

```bash
mysql -u root -h 127.0.0.1 --port=33060 elephant
```

# Updating [pkg.go.dev](https://pkg.go.dev/github.com/gonimals/elephant)

These commands update the golang servers, so go get -u in other repos pull the latest commit too

```bash
bash  # create a subshell to write environment variables

export TZ=UTC0
TIMESTAMP=$(git log -1 --date=format-local:%Y%m%d%H%M%S --pretty=format:"%cd")
HASH=$(git rev-parse --short=12 HEAD)
export GOPROXY=https://proxy.golang.org
go list -m "github.com/gonimals/elephant@v0.0.0-$TIMESTAMP-$HASH"

exit
```