# dbd

Build command (linux only - can't cross compile cgo used by mattn/go-sqlite3):
$go build -a -v -tags netgo -ldflags '-extldflags "-lm -lstdc++ -static"' .
