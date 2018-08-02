package conf

import (
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/topology"
	"path/filepath"
)

const (
	ErrorAddr = "Unable to load addresses"
	ErrorTopo = "Unable to load topology"
)

type Conf struct {
	// ID is the element ID
	ID string
	// Topo contains the name of all local infra, a map of interface IDs to routers,
	// and the actual topology
	Topo *topology.Topo
	// ConfDir is the configuration directory
	ConfDir string
	// BindAddr is the local bind address
	BindAddr *snet.Addr
	// PublicAddr is the public address
	PublicAddr *snet.Addr
}

// Load initalizes the configuration by loading it from confDir
func Load(id string, confDir string) (*Conf, error) {
	c := &Conf{
		ID:      id,
		ConfDir: confDir,
	}
	if err := c.loadTopo(); err != nil {
		return nil, err
	}
	return c, nil
}
func (c *Conf) loadTopo() error {
	var err error
	path := filepath.Join(c.ConfDir, topology.CfgName)
	if c.Topo, err = topology.LoadFromFile(path); err != nil {
		return common.NewBasicError(ErrorTopo, nil)
	}
	// load public and bind address
	topoAddr, ok := c.Topo.AP[c.ID]
	if !ok {
		return common.NewBasicError(ErrorAddr, nil, "err", "Element ID not found", "id", c.ID)
	}
	publicInfo := topoAddr.PublicAddrInfo(c.Topo.Overlay)
	c.PublicAddr = &snet.Addr{IA: c.Topo.ISD_AS, Host: addr.HostFromIP(publicInfo.IP),
		L4Port: uint16(publicInfo.L4Port)}
	bindInfo := topoAddr.BindAddrInfo(c.Topo.Overlay)
	tmpBind := &snet.Addr{IA: c.Topo.ISD_AS, Host: addr.HostFromIP(bindInfo.IP),
		L4Port: uint16(bindInfo.L4Port)}
	if !tmpBind.EqAddr(c.PublicAddr) {
		c.BindAddr = tmpBind
	}
	return nil
}
