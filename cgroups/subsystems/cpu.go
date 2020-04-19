package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpuSubsystem struct {

}

func (cs *CpuSubsystem) Name() string {
	return "cpu"
}

func (cs *CpuSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsystemCgroupPath, err := GetCgroupPath(cs.Name(), cgroupPath, true); err != nil {
		return err
	} else {
		if res.CpuShare != "" {
			if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpu.shares"),
				[]byte(res.CpuSet), 0644); err != nil {
				return fmt.Errorf ("set cgroup cpu share fail %v", err)
			}
		}
		return nil
	}
}

func (cs *CpuSubsystem) Apply(cgroupPath string, pid int) error {
	if subsystemCgroupPath, err := GetCgroupPath(cs.Name(), cgroupPath, false); err != nil {
		return fmt.Errorf ("get cgroup %v error: %v", cgroupPath , err)
	} else {
		if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf ("apply cgroup proc fail %v", err)
		}
		return nil
	}
}

func (cs *CpuSubsystem) Remove(cgroupPath string) error {
	if subsystemCgroupPath, err := GetCgroupPath(cs.Name(), cgroupPath, false); err != nil {
		return fmt.Errorf ("get cgroup %v error: %v", cgroupPath , err)
	} else {
		return os.RemoveAll(subsystemCgroupPath)
	}
}
