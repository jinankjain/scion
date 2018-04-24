package main

// #include <stdint.h>
// #include <sys/socket.h>
// #include <zlog.h>
// #include <scion/scion.h>
// #include <tcp/middleware.h>
// #include <uthash.h>
import "C"

import (
	"os"
	"fmt"
	"syscall"
	"unsafe"
)

type zlogCategory C.struct_zlog_category_t
type chkInput C.struct_chk_input

var zc *zlogCategory
var chkUdpInput *chkInput
var app_socket int
var data_v4_socket int
var data_v6_socket int
var sockPath string

const (
	MAX_BACKLOG = 5
)

func main() {

	err := os.Setenv("TZ", "UTC")
	if err != nil {
		fmt.Println("Unable to set environment variable TZ")
		os.Exit(-1)
	}
	zlogCfg := os.Getenv("ZLOG_CFG")
	if zlogCfg == "" {
		zlogCfg = "c/dispatcher/dispatcher.conf"
	}
	if C.zlog_init(C.CString(zlogCfg)) < 0 {
		fmt.Println("failed to init zlog cfg ", zlogCfg)
		os.Exit(-1)
	}
	zc = (*zlogCategory)(unsafe.Pointer(C.zlog_get_category(C.CString("dispatcher"))))
	if zc == nil {
		fmt.Println("failed to get dispatcher zlog category")
		C.zlog_fini()
		os.Exit(-1)
	}

	// TCPMW uses many open files, set rlimit as high as possible
	var rlim syscall.Rlimit
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		fmt.Println("getrlimit(): ", err)
		os.Exit(-1)
	}
	fmt.Printf("Changing RLIMIT_NOFILE %d -> %d\n", rlim.Cur, rlim.Max)
	rlim.Cur = rlim.Max
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		fmt.Println("setrlimit(): ", err)
		os.Exit(-1)
	}

	// Allocate for later use
	chkUdpInput = (*chkInput)(unsafe.Pointer(C.mk_chk_input(C.UDP_CHK_INPUT_SIZE)))

	if err = createSockets(); err != nil {
		fmt.Println(err)
	}
}

func createSockets() error {
	var err error
	app_socket, err = syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	data_v4_socket, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return err
	}
	data_v6_socket, err = syscall.Socket(syscall.AF_INET6, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return err
	}

	if err = setSocketopts(); err != nil {
		return err
	}
	if err = bindAppSocket(); err != nil {
		return err
	}
	return nil
}

func setSocketopts() error {
	var optval int
	var err error
	// FIXME(kormat): This should go away once the dispatcher and the router no
	// longer try binding to the same socket
	if data_v4_socket > 0 {
		if err = syscall.SetsockoptInt(data_v4_socket, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, optval); err != nil {
			return err
		}
		if err = syscall.SetsockoptInt(data_v4_socket, syscall.SOL_SOCKET, syscall.SO_RXQ_OVFL, optval); err != nil {
			return err
		}
		if err = syscall.SetsockoptInt(data_v4_socket, syscall.IPPROTO_IP, syscall.IP_PKTINFO, optval); err != nil {
			return err
		}
	}
	if data_v6_socket > 0 {
		if err = syscall.SetsockoptInt(data_v6_socket, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, optval); err != nil {
			return err
		}
		if err = syscall.SetsockoptInt(data_v6_socket, syscall.SOL_SOCKET, syscall.SO_RXQ_OVFL, optval); err != nil {
			return err
		}
		if err = syscall.SetsockoptInt(data_v6_socket, syscall.IPPROTO_IPV6, syscall.IPV6_RECVPKTINFO, optval); err != nil {
			return err
		}
		if err = syscall.SetsockoptInt(data_v6_socket, syscall.SOL_IPV6, syscall.IPV6_V6ONLY, optval); err != nil {
			return err
		}
	}
	optval = 1 << 20
	if err = syscall.SetNonblock(app_socket, true); err != nil {
		return err
	}
	if data_v4_socket > 0 {
		if err = syscall.SetsockoptInt(data_v4_socket, syscall.SOL_SOCKET, syscall.SO_RCVBUF, optval); err != nil {
			return err
		}
		if err = syscall.SetNonblock(data_v4_socket, true); err != nil {
			return err
		}
	}
	if (data_v6_socket > 0) {
		if err = syscall.SetsockoptInt(data_v6_socket, syscall.SOL_SOCKET, syscall.SO_RCVBUF, optval); err != nil {
			return err
		}
		if err = syscall.SetNonblock(data_v6_socket, true); err != nil {
			return err
		}
	}
	return nil
}

func bindAppSocket() error {
	var err error
	var su syscall.SockaddrUnix
	env := os.Getenv("DISPATCHER_ID")
	if env == "" {
		env = C.DEFAULT_DISPATCHER_ID
	}
	sockPath = fmt.Sprintf("%s/%s.sock", C.DISPATCHER_DIR, env)
	su.Name = sockPath
	if err = os.Mkdir(C.DISPATCHER_DIR, 0755); err != nil {
		return err
	}
	if err = syscall.Bind(app_socket, &su); err != nil {
		return err
	}
	if err = syscall.Listen(app_socket, MAX_BACKLOG); err != nil {
		return err
	}
	return nil
}

func bindDataSockets() error {
	var err error
	var sa *C.struct_sockaddr_storage

	if data_v4_socket > 0 {
		var sin syscall.SockaddrInet4
		sin.Addr = [4]byte{0, 0, 0, 0}
		sin.Port = C.SCION_UDP_EH_DATA_PORT
		if err = syscall.Bind(data_v4_socket, &sin); err != nil {
			return err
		}
	}
	return nil
}
