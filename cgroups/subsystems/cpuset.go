package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpusetSubsystem struct {

}

func (css *CpusetSubsystem) Name() string {
	return "cpusset"
}

func (css *CpusetSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsystemCgroupPath, err := GetCgroupPath(css.Name(), cgroupPath, true); err != nil {
		return err
	} else {
		if res.CpuSet != "" {
			if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpuset.cpus"),
				[]byte(res.CpuSet), 0644); err != nil {
				return fmt.Errorf ("set cgroup cpuset fail %v", err)
			}
		}
		return nil
	}
}

func (css *CpusetSubsystem) Apply(cgroupPath string, pid int) error {
	if subsystemCgroupPath, err := GetCgroupPath(css.Name(), cgroupPath, false); err != nil {
		return fmt.Errorf ("get cgroup %v error: %v", cgroupPath , err)
	} else {
		if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf ("apply cgroup proc fail %v", err)
		}
		return nil
	}
}

func (css *CpusetSubsystem) Remove(cgroupPath string) error {
	if subsystemCgroupPath, err := GetCgroupPath(css.Name(), cgroupPath, false); err != nil {
		return fmt.Errorf ("get cgroup %v error: %v", cgroupPath , err)
	} else {
		return os.RemoveAll(subsystemCgroupPath)
	}
}
