@0xff9c39d647ba1420;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

struct APNADMsg {
    id @0 :UInt64; #Request ID
    union {
        ephIDGenerationReq @1 :EphIDGenerationReq;
        ephIDGenerationReply @2 :EphIDGenerationReply;
        dNSReq @3 :DNSReq;
        dNSReply @4: DNSReply;
    }
}

struct EphIDGenerationReq {
    kind @0: UInt8;
    addr @1: ServiceAddr;
    pubkey @2: Data;
    server @3: UInt8;
}

struct DNSReq {
    addr @0: ServiceAddr;
}

struct DNSReply {
    errorCode @0: UInt8;
    certificate @1: Certificate;
}

struct Certificate {
    ephid @0: Data;
    pubkey @1: Data;
    recvOnly @2: Bool;
    expTime @3: Data;
    signature @4: Data;
}

struct EphIDGenerationReply {
    errorCode @0: UInt8;
    ephid @1: Data;
}

struct ServiceAddr {
    addr @0 :Data;
    protocol @1: UInt8;
}
