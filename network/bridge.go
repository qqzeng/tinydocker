package network

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
	"time"
)

type BridgeNetworkDriver struct {
}

func (bnd *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (bnd *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	cidr, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("fail to parse subnet %s : %v", name, err)
	}
	ipNet.IP = cidr
	nw := &Network{
		Name:    name,
		IpRange: ipNet,
		Driver: bnd.Name(),
	}
	err = bnd.initBridge(nw)
	if err != nil {
		log.Errorf("Fail to initialize bridge network : %v", err)
	}
	return nw, nil
}

func (bnd *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	iface, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("fail to get interface %s : %v", bridgeName, err)
	}
	if err := netlink.LinkDel(iface); err != nil {
		return fmt.Errorf("fail to remove bridge interface %s delete: %v", bridgeName, err)
	}
	return nil
}

func (bnd *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	bridgeName := network.Name
	iface, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("fail to get interface %s : %v", bridgeName, err)
	}
	la := netlink.NewLinkAttrs()
	la.Name = endpoint.Id[:5]
	la.MasterIndex = iface.Attrs().Index
	endpoint.Device = netlink.Veth {
		LinkAttrs: la,
		PeerName: "cif-" + endpoint.Id[:5],
	}
	if err := netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("fail to add endpoint interface : %v", err)
	}
	if err := netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("fail start up endpoint interface : %v", err)
	}

	return nil
}

func (bnd *BridgeNetworkDriver) Disconnect(network *Network, endpoint *Endpoint) error {
	return nil
}

func (bnd *BridgeNetworkDriver) initBridge(network *Network) error {
	/* 1. create virtual device bridge */
	bridgeName := network.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("fail to add bridge network %s : %v", bridgeName, err)
	}
	/* 2. set ip address and route for bridge */
	gatewapIp := *network.IpRange
	gatewapIp.IP = network.IpRange.IP
	if err := setInterfaceIp(bridgeName, gatewapIp.String()); err != nil {
		return fmt.Errorf("fail to assign address %s to bridge %s : %v", gatewapIp, bridgeName, err)
	}
	/* 3. start up bridge */
	if err := setInterfaceUp(bridgeName); err != nil {
		return fmt.Errorf("fail to start up bridge %s : %v", bridgeName, err)
	}
	/* 4. set SNAT rules of iptables for container and host */
	if err := setIptables(bridgeName, network.IpRange); err != nil {
		return fmt.Errorf("fail to set iptables for bridge %s : %v", bridgeName, err)
	}
	return nil
}

func createBridgeInterface(bridgeName string) error {
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	br := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("fail to create bridge %s : %v", bridgeName, err)
	}

	/* associate host eth and bridge */
	hostEthName := "eth0"
	hostEth, _ := netlink.LinkByName(hostEthName)
	if err := netlink.LinkSetMaster(hostEth, br); err != nil {
		return fmt.Errorf("fail to associate host eth %s and bridge: %v", hostEthName, err)
	}
	return nil
}

func setInterfaceIp(bridgeName string, rawIp string) error {
	//iface, err := netlink.LinkByName(bridgeName)
	//if err != nil {
	//	return fmt.Errorf("fail to get interface %s : %v", bridgeName, err)
	//}
	retries := 2
	var iface netlink.Link
	var err error
	for i := 0; i < retries; i++ {
		iface, err = netlink.LinkByName(bridgeName)
		if err == nil {
			break
		}
		log.Debugf("error retrieving new bridge netlink link [ %s ]... retrying", bridgeName)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("abandoning retrieving the new bridge link from netlink, Run [ ip link ] to troubleshoot the error: %v", err)
	}
	ipNet, err := netlink.ParseIPNet(rawIp)
	if err != nil {
		return fmt.Errorf("fail to parse ip address %s : %v", rawIp, err)
	}
	addr := &netlink.Addr{IPNet: ipNet, Peer: ipNet, Label: "", Flags: 0, Scope: 0}
	return netlink.AddrAdd(iface, addr)
}

func setInterfaceUp(bridgeName string) error {
	iface, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("fail to get interface %s : %v", bridgeName, err)
	}
	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("fail to enable interface %s : %v", bridgeName, err)
	}
	return nil
}

func setIptables(bridgeName string, subnet *net.IPNet) error {
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		log.Errorf("output : %s, fail to execute iptables %v", output, err)
	}
	return nil
}

