#!/bin/bash

cmd_run_apna_server() {
    echo "Running apna server"
    ./bin/apna_app -a go/examples/apna_app/testdata/server.json -l 2-ff00:0:222,[127.0.0.1]:50049 server
}

cmd_run_apna_client() {
    echo "Running apna client"
    ./bin/apna_app -a go/examples/apna_app/testdata/client.json -r 2-ff00:0:222,[127.0.0.1]:50049 -l 1-ff00:0:133,[127.0.0.4]:50045 client
}

cmd_run_apna_pslabserver() {
    echo "Running apna server"
    ./bin/apna_app -a go/examples/apna_app/testdata/server_slab.json -l 19-ffaa:1:d3,[127.0.0.1]:50049 server
}

cmd_run_apna_slabclient2() {
    echo "Running apna client"
    ./bin/apna_app -a go/examples/apna_app/testdata/client_slab.json -r 20-ffaa:1:d4,[127.0.0.1]:50049 -l 19-ffaa:1:d3,[127.0.0.4]:50045 client
}


cmd_run_apna_slabclient1() {
    echo "Running apna client"
    ./bin/apna_app -a go/examples/apna_app/testdata/client_slab.json -r 17-ffaa:1:d5,[127.0.0.1]:50049 -l 19-ffaa:1:d3,[127.0.0.4]:50045 client
}

cmd_run_ms() {
    echo "Running management service"
    ./bin/apna_ms -config go/apna_ms/testdata/apna_ms.json
}

COMMAND="$1"

case "$COMMAND" in
    server) cmd_run_apna_server ;;
    client) cmd_run_apna_client ;;
    psserver) cmd_run_apna_pslabserver ;;
    sclient1) cmd_run_apna_slabclient1 ;;
    sclient2) cmd_run_apna_slabclient1 ;;
    ms) cmd_run_ms ;;
esac

