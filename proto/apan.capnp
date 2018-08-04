@0xf8dd029a7e8e946e;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

struct APNAPkt {
    localEphID @0: Data; #Local EphID
    remoteEphID @1: Data; #Remote EphID
    remotePort @2: UInt16; #Remote Port
    localPort @3: UInt16; #LocalPort
    packetMAC @4: Data; #Packet MAC
    nextHeader @5: UInt8;
    union {
        pubkey @6: Data;
        ecert @7: Data;
        data @8: Data;
        ecertPubkey @9: EcertPubkey;
    }
}

struct EcertPubkey {
    ecert @0: Data;
    pubkey @1: Data;
}
