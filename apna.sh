#!/bin/bash

cmd_run_apna_server() {
    echo "Running apna server"
    ./bin/apnaapp -a go/examples/apnaapp/testdata/server.json -l 2-ff00:0:222,[127.0.0.1]:50049 server
}

cmd_run_apna_client() {
    echo "Running apna client"
    ./bin/apnaapp -a go/examples/apnaapp/testdata/client.json -r 2-ff00:0:222,[127.0.0.1]:50049 -l 1-ff00:0:133,[127.0.0.4]:50048 client
}

cmd_run_ms() {
    echo "Running management service"
    ./bin/apnad -config go/apnad/testdata/apnad.json
}

COMMAND="$1"

case "$COMMAND" in
    server) cmd_run_apna_server ;;
    client) cmd_run_apna_client ;;
    ms) cmd_run_ms ;;
esac
