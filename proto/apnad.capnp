@0xff9c39d647ba1420;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

struct APNADMsg {
    id @0 :UInt64; #Request ID
    union {
        ephidGenerationReq @1 :EphIDGenerationReq;
        ephidGenerationReply @2 :EphIDGenerationReply;
        dnsReq @3 :DNSReq;
        dnsReply @4: DNSReply;
    }
}

struct EphIDGenerationReq {
    kind @0: UInt8;
    addr @1: ServiceAddr;
}

struct DNSReq {
    addr @0: ServiceAddr;
}

struct DNSReply {
    errorCode @0: UInt8;
    ephid @1: Data;
}

struct EphIDGenerationReply {
    errorCode @0: UInt8;
    ephid @1: Data;
}

struct ServiceAddr {
    addr @0 :Data;
    protocol @1: UInt8;
}