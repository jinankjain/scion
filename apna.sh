#!/bin/bash

cmd_run_apna_server() {
    echo "Running apna server"
    ./bin/apna_app -a go/examples/apna_app/testdata/server.json -l 2-ff00:0:222,[127.0.0.1]:50049 server
}

cmd_run_apna_client() {
    echo "Running apna client"
    ./bin/apna_app -a go/examples/apna_app/testdata/client.json -r 2-ff00:0:222,[127.0.0.1]:50049 -l 1-ff00:0:133,[127.0.0.4]:50048 client
}

cmd_run_ms() {
    echo "Running management service"
    ./bin/apna_ms -config go/apna_ms/testdata/apna_ms.json
}

COMMAND="$1"

case "$COMMAND" in
    server) cmd_run_apna_server ;;
    client) cmd_run_apna_client ;;
    ms) cmd_run_ms ;;
esac

