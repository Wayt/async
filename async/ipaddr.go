package async

import (
	"net"

	"github.com/juju/errors"
)

var (
	ErrNoIPv4Found         = errors.New("cannot find ipv4 address")
	ErrInterfaceNotUP      = errors.New("interface is not UP")
	ErrInterfaceIsLoopback = errors.New("interface is loopback")
)

func loadInterfaceIPv4(name string) (net.IP, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, errors.Annotatef(err, "could not load interface ipv4")
	}

	if iface.Flags&net.FlagUp == 0 {
		return nil, ErrInterfaceNotUP
	}

	if iface.Flags&net.FlagLoopback != 0 {
		return nil, ErrInterfaceIsLoopback
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	for _, rawAddr := range addrs {
		var ip net.IP
		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}

		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4, nil
		}
	}

	return nil, ErrNoIPv4Found
}
