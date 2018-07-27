@0xb88e42fba15deb5d;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

struct APNAHeader {
    localEphID @0: Data; #Local EphID
    remoteEphID @1: Data; #Remote EphID
    nextHeader @2: UInt8;
    union {
        pubkey @3: Data;
        ecert @4: Data;
        data @5: Data;
        ecertPubkey @6: EcertPubkey;
    }
}

struct EcertPubkey {
    ecert @0: Data;
    pubkey @1: Data;
}
