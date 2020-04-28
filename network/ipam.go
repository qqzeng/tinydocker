package network

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
)

type IPAM struct {
	/* ip allocation information saved path */
	SubnetAllocatorPath string
	/* key->Range of Ip, value->bitmap array */
	Subnets *map[string]string
}

const IPAMDefaultAllocatorPath = "/var/run/tinydocker/network/ipam/subnet.json"

var ipAllocator = &IPAM{
	SubnetAllocatorPath: IPAMDefaultAllocatorPath,
}

/* allocate an available ip address from a specified range of ip address */
func (ipam *IPAM) Allocate(subnet *net.IPNet) (net.IP, error) {
	ipam.Subnets = &map[string]string{}
	err := ipam.load()
	if err != nil {
		log.Error(err)
	}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	one, bits := subnet.Mask.Size()
	/* not allocated already, initialize it  */
	if _, exists := (*ipam.Subnets)[subnet.String()]; !exists {
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(bits-one))
	}
	/* find the first unallocated bit, and update it  */
	var ip net.IP
	for c := range (*ipam.Subnets)[subnet.String()] {
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			ipAllocBytes := []byte((*ipam.Subnets)[subnet.String()])
			ipAllocBytes[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipAllocBytes)
			ip = subnet.IP
			for t := uint(4); t > 0; t-=1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			ip[3] += 1
			break
		}
	}
	err = ipam.dump()
	return ip, err
}

/* free an unused ip address to a specified range of ip address */
func (ipam *IPAM) Release(subnet *net.IPNet, ipAddr *net.IP) error {
	ipam.Subnets = &map[string]string{}
	err := ipam.load()
	if err != nil {
		log.Error(err)
	}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	c := 0
	releasedIp := ipAddr.To4()
	releasedIp[3] -= 1
	for t := uint(4); t > 0; t-=1 {
		c += int(releasedIp[t-1] - subnet.IP[t-1]) << ((4 - t) * 8)
	}
	ipAllocBytes := []byte((*ipam.Subnets)[subnet.String()])
	ipAllocBytes[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipAllocBytes)
	err = ipam.dump()
	return err
}

/* load ip allocation information from default saved file */
func (ipam *IPAM) load() error {
	allocatorPath := ipam.SubnetAllocatorPath
	exists, err := PathExists(allocatorPath)
	if err != nil {
		return fmt.Errorf("fail to check ipam saved directory %s existence : %v", allocatorPath, err)
	}
	if !exists {
		log.Infof("fail to retrive ipam saved directory %s : %v", allocatorPath, err)
		return nil
	}
	ipamFile, err :=  os.Open(allocatorPath)
	defer ipamFile.Close()

	if err != nil {
		return fmt.Errorf("fail to open ipam saved file %s : %v", allocatorPath, err)
	}
	var ipamBytes []byte
	ipamBytes, err = ioutil.ReadAll(ipamFile)
	if err != nil {
		return fmt.Errorf("fail to read ipam saved file %s : %v", allocatorPath, err)
	}
	err = json.Unmarshal(ipamBytes, ipam)
	if err != nil {
		return fmt.Errorf("fail to unmarshal ipam : %v", err)
	}
	return nil
}

/* dumped allocation information to default saved file */
func (ipam *IPAM) dump() error {
	dumpFilePath := ipam.SubnetAllocatorPath
	dumpDir, _:=  path.Split (dumpFilePath)
	exists, err := PathExists(dumpDir)
	if err != nil {
		return fmt.Errorf("fail to check ipam dumpped %s existence : %v", dumpDir, err)
	}
	if !exists {
		if err := os.MkdirAll(dumpDir, 0644); err != nil {
			return fmt.Errorf("fail to create ipam dumpped path %s : %v", dumpDir, err)
		}
	}
	ipamFile, err := os.OpenFile(dumpFilePath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("fail to create ipam dumpped file %s : %v", dumpFilePath, err)
	}
	defer ipamFile.Close()
	ipamBytes, err := json.Marshal(ipam)
	if err != nil {
		return fmt.Errorf("fail to marshal ipam : %v", err)
	}
	_, err = ipamFile.Write(ipamBytes)
	if err != nil {
		return fmt.Errorf("fail to write ipam file %s : %v", dumpFilePath, err)
	}
	return nil
}

