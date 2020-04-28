package network

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"github.com/qqzeng/tinydocker/container"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
)

var (
	defaultNetworkPath = "/var/run/tinydocker/network/network/"
	drivers = map[string]NetworkDriver{}
	networks = map[string]*Network{}
)

type Network struct {
	Name string			/* name of network */
	IpRange *net.IPNet	/* the range of ip address of network */
	Driver string 		/* driver name of network */
}

type Endpoint struct {
	Id	string	`json:id`						/* id of endpoint */
	Device	netlink.Veth `json:"dev"`			/* device of veth of endpoint */
	IPAddress net.IP `json:"ip"`				/* ip address of endpoint */
	MacAddress net.HardwareAddr `json:"mac"`	/* mac address of device of endpoint */
	PortMapping []string `json:"portmapping"`	/* port mapping of endpoint */
	Network *Network							/* connected network of endpoint */
}

type NetworkDriver interface {
	/* driver name */
	Name() string
	/* create a network with specific range of ip address and name */
	Create(subnet string, name string) (*Network, error)
	/* delete a network */
	Delete(network Network) error
	/* connect an endpoint of network to a network */
	Connect(network *Network, endpoint *Endpoint) error
	/* disconnect an endpoint of network from a network */
	Disconnect(network *Network, endpoint *Endpoint) error
}

/* create a network */
func CreateNetwork(driver, subnet, name string) error {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("fail to parse cidr ip address %s : %v", subnet, err)
	}
	ip, err := ipAllocator.Allocate(ipNet)
	if err != nil {
		return fmt.Errorf("fail to allocate cidr ip address %s : %v", subnet, err)
	}
	ipNet.IP = ip
	nw, err := drivers[driver].Create(ipNet.String(), name)
	if err != nil {
		return fmt.Errorf("fail to create network for cidr ip address %s : %v", subnet, err)
	}
	return nw.dump(defaultNetworkPath)
}

func Connect(nwName string, cInfo *container.ContainerInfo) error {
	nw, ok := networks[nwName]
	if !ok {
		return fmt.Errorf("fail to retrive network %s", nwName)
	}
	ip, err := ipAllocator.Allocate(nw.IpRange)
	if err != nil {
		return fmt.Errorf("fail to allocate ip address for network %s : %v", nwName, err)
	}
	ep := &Endpoint{
		Id:          fmt.Sprintf("%s-%s", cInfo.Id, nwName),
		IPAddress:   ip,
		PortMapping: cInfo.PortMapping,
		Network:     nw,
	}
	if err := drivers[nw.Driver].Connect(nw, ep); err != nil {
		return fmt.Errorf("fail to connect endpoint to network %s : %v", nwName, err)
	}
	if err := configEndpointIpAddressAndRoute(ep, cInfo); err != nil {
		return fmt.Errorf("fail to configure ip address for endpoint : %v", err)
	}
	if err := configPortMapping(ep, cInfo); err != nil {
		return fmt.Errorf("fail to configure port mapping for endpoint and network : %v", err)
	}
	return nil
}

func Disconnect(nwName string, cInfo *container.ContainerInfo) error {
	return nil
}

/* load  and populate network information list */
func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver
	exists, err := PathExists(defaultNetworkPath)
	if err != nil {
		return fmt.Errorf("fail to check network dumped directory %s existence : %v", defaultNetworkPath, err)
	}
	if !exists {
		if err := os.MkdirAll(defaultNetworkPath, 0644); err != nil {
			return fmt.Errorf("fail to create network dumped directory %s : %v", defaultNetworkPath, err)
		}
	}
	filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name : nwName,
		}
		if err := nw.load(nwPath); err != nil {
			log.Errorf("fail to load network information : %v", err)
		}
		networks[nwName] = nw
		return nil
	})
	return nil
}

func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12 ,  1,  3,' ', 0)
	fmt.Fprintf(w, "Name\tIpRange\tDriver\n")
	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n", nw.Name, nw.IpRange.String(), nw.Driver)
	}
	if err := w.Flush(); err != nil {
		log.Errorf("Write content information to stdout error : %v", err)
	}
}

func DeleteNetwork(nwName string) error {
	nw, ok := networks[nwName]
	if !ok {
		return fmt.Errorf("fail to retrive network %s", nwName)
	}
	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("fail to remove network %s gateway %v", nwName, err)
	}
	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("fail to remove network %s : %v", nwName, err)
	}
	return nw.remove()
}

func (nw *Network) remove() error {
	dumpFilePath := path.Join(defaultNetworkPath, nw.Name)
	exists, err := PathExists(dumpFilePath)
	if err != nil {
		return fmt.Errorf("fail to check dump directory %s existence : %v", dumpFilePath, err)
	}
	if !exists {
		return nil
	}
	return os.Remove(dumpFilePath)
}

/* create file and save network content to it */
func (nw *Network) dump(dumpDir string) error {
	exists, err := PathExists(dumpDir)
	if err != nil {
		return fmt.Errorf("fail to check dump directory %s existence : %v", dumpDir, err)
	}
	if !exists {
		if err := os.MkdirAll(dumpDir, 0644); err != nil {
			return fmt.Errorf("fail to create dump directory %s : %v", dumpDir, err)
		}
	}
	nwDir := path.Join(dumpDir, nw.Name)
	nwFile, err := os.OpenFile(nwDir, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("fail to create dump file %s : %v", nwDir, err)
	}
	defer nwFile.Close()
	nwBytes, err := json.Marshal(nw)
	if err != nil {
		return fmt.Errorf("fail to marshal network : %v", err)
	}
	_, err = nwFile.Write(nwBytes)
	if err != nil {
		return fmt.Errorf("fail to write network file %s : %v", nwDir, err)
	}
	return nil
}

/* load network information from file */
func (nw *Network) load(dumpFilePath string) error {
	exists, err := PathExists(dumpFilePath)
	if err != nil {
		return fmt.Errorf("fail to check dump file path %s existence : %v", dumpFilePath, err)
	}
	if !exists {
		return fmt.Errorf("network dumped file %s does not exist : %v", dumpFilePath, err)
	}
	nwFile, err :=  os.Open(dumpFilePath)
	if err != nil {
		return fmt.Errorf("fail to open network dumped file %s : %v", dumpFilePath, err)
	}
	var nwBytes []byte
	nwBytes, err = ioutil.ReadAll(nwFile)
	if err != nil {
		return fmt.Errorf("fail to read network dumped file %s : %v", dumpFilePath, err)
	}
	err = json.Unmarshal(nwBytes, nw)
	if err != nil {
		return fmt.Errorf("fail to unmarshal network : %v", err)
	}
	return nil
}

func PathExists(url string) (bool, error) {
	_, err := os.Stat(url)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}


func configEndpointIpAddressAndRoute(ep *Endpoint,  cInfo *container.ContainerInfo) error {
	peerName := ep.Device.PeerName
	peerLink, err := netlink.LinkByName(peerName)
	if err != nil {
		return fmt.Errorf("fail to configure endpoint: %v", err)
	}
	defer enterContainerNetNs(&peerLink, cInfo)()
	interfaceIp := *ep.Network.IpRange
	interfaceIp.IP = ep.IPAddress
	if err := setInterfaceIp(peerName, interfaceIp.String()); err != nil {
		return fmt.Errorf("fail to configure endpoint %s's ip: %v", ep.Network, err)
	}
	if err = setInterfaceUp(peerName); err != nil {
		return fmt.Errorf("fail to enable endpoint %s : %v", peerName, err)
	}
	if err = setInterfaceUp("lo"); err != nil {
		return fmt.Errorf("fail to enable endpoint loopback : %v", err)
	}
	_, ipNet, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw: ep.Network.IpRange.IP,
		Dst: ipNet,
	}
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("fail to configure default route : %v", err)
	}
	return nil
}

func enterContainerNetNs(enLink *netlink.Link,  cInfo *container.ContainerInfo) func()  {
	netFile, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cInfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("fail to get network namespace of container process : %v", err)
	}
	nsFd := netFile.Fd()
	runtime.LockOSThread()
	if err := netlink.LinkSetNsFd(*enLink, int(nsFd)); err != nil {
		log.Errorf("fail to set link namespace : %v", err)
	}
	curNs, err := netns.Get()
	if err != nil {
		log.Errorf("fail to get current namespace : %v", err)
	}
	if err := netns.Set(netns.NsHandle(nsFd)); err != nil {
		log.Errorf("fail to set namespace : %v", err)
	}
	return func() {
		netns.Set(curNs)
		curNs.Close()
		runtime.UnlockOSThread()
		netFile.Close()
	}
}

func configPortMapping(ep *Endpoint, cInfo *container.ContainerInfo) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			return fmt.Errorf("fail to parse portmapping array : %s\n", pm)
		}
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("output : %s, fail to execute iptables %v", output, err)
		}
	}
	return nil
}
