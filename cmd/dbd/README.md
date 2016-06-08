# dbd

Build command:
$ GOOS=linux go build -a -v -tags netgo -ldflags '-extldflags "-lm -lstdc++ -static"' .
