@0xb88e42fba15deb5d;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

struct APNAHeader {
    localEphID @0: Data; #Local EphID
    remoteEphID @1: Data; #Remote EphID
    packetMAC @2: Data; #Packet MAC
    nextHeader @3: UInt8;
    union {
        pubkey @4: Data;
        ecert @5: Data;
        data @6: Data;
        ecertPubkey @7: EcertPubkey;
    }
}

struct EcertPubkey {
    ecert @0: Data;
    pubkey @1: Data;
}
