package transport

import (
	"context"
	"net"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/util"
)

var _ infra.Transport = (*UDPTransport)(nil)

// UDPTransport implements interface Transport by wrapping around a
// net.PacketConn. The reliability of the underlying net.PacketConn defines the
// semantics behind SendMsgTo and SendUnreliableMsgTo.
//
// For PacketTransports running on top of UDP, both SendMsgTo and
// SendUnreliableMsgTo are unreliable.
//
// For PacketTransports running on top of UNIX domain socket with SOCK_DGRAM or
// Reliable socket, both SendMsgTo and SendUnreliableMsgTo guarantee reliable
// delivery to the other other end of the socket. Note that in this case, the
// reliability only extends to the guarantee that the message was not lost in
// transfer. It is not a guarantee that the server has read and processed the
// message.
type UDPTransport struct {
	conn *net.UDPConn
	// While conn is safe for use from multiple goroutines, deadlines are
	// global so it is not safe to enforce two at the same time. Thus, to
	// meet context deadlines we serialize access to the conn.
	writeLock *util.ChannelLock
	readLock  *util.ChannelLock
}

func NewUDPTransport(conn *net.UDPConn) *UDPTransport {
	return &UDPTransport{
		conn:      conn,
		writeLock: util.NewChannelLock(),
		readLock:  util.NewChannelLock(),
	}
}

func (u *UDPTransport) SendUnreliableMsgTo(ctx context.Context, b common.RawBytes,
	address net.Addr) error {

	select {
	case <-u.writeLock.Lock():
		defer u.writeLock.Unlock()
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := setWriteDeadlineFromCtx(u.conn, ctx); err != nil {
		return err
	}
	n, err := u.conn.Write(b)
	if n != len(b) {
		return common.NewBasicError("Wrote incomplete message", nil, "wrote", n, "expected", len(b))
	}
	return err
}

func (u *UDPTransport) SendMsgTo(ctx context.Context, b common.RawBytes,
	address net.Addr) error {

	return u.SendUnreliableMsgTo(ctx, b, address)
}

func (u *UDPTransport) RecvFrom(ctx context.Context) (common.RawBytes, net.Addr, error) {
	select {
	case <-u.readLock.Lock():
		defer u.readLock.Unlock()
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
	if err := setReadDeadlineFromCtx(u.conn, ctx); err != nil {
		return nil, nil, err
	}
	b := make(common.RawBytes, common.MaxMTU)
	n, address, err := u.conn.ReadFromUDP(b)
	return b[:n], address, err
}

func (u *UDPTransport) Close(context.Context) error {
	return u.conn.Close()
}
