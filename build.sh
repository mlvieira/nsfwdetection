#!/bin/env bash

# set tensorflor include and library path for tsl header
export CGO_CFLAGS="-I/usr/local/include/tensorflow"
export CGO_LDFLAGS="-L/usr/local/lib"

go build -o dist/detectnsfw cmd/server/*.go
