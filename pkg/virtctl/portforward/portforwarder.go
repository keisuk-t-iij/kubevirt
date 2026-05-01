package portforward

import (
	"errors"
	"net"
	"strings"

	v1 "kubevirt.io/api/core/v1"
	kvcorev1 "kubevirt.io/client-go/kubevirt/typed/core/v1"
	"kubevirt.io/client-go/log"
)

type portForwarder struct {
	kind, namespace, name string
	resource              portforwardableResource
}

type portforwardableResource interface {
	PortForward(name string, port int, protocol string) (kvcorev1.StreamInterface, error)
}

type vsockableResource interface {
	VSOCK(name string, options *v1.VSOCKOptions) (kvcorev1.StreamInterface, error)
}

type vsockResourceAdapter struct {
	resource vsockableResource
}

func (a *vsockResourceAdapter) PortForward(name string, port int, _ string) (kvcorev1.StreamInterface, error) {
	return a.resource.VSOCK(name, &v1.VSOCKOptions{TargetPort: uint32(port)})
}

func (p *portForwarder) startForwarding(address *net.IPAddr, port forwardedPort) error {
	log.Log.Infof("forwarding %s %s:%d to %d", port.protocol, address, port.local, port.remote)
	if port.protocol == protocolUDP {
		return p.startForwardingUDP(address, port)
	}

	if port.protocol == protocolTCP {
		return p.startForwardingTCP(address, port)
	}

	return errors.New("unknown protocol: " + port.protocol)
}

func handleConnectionError(err error, port forwardedPort) {
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		log.Log.Errorf("error handling connection for %d: %v", port.local, err)
	}
}
