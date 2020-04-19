package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type MemorySubsystem struct {
}

func (ms *MemorySubsystem) Name() string {
	return "memory"
}

func (ms *MemorySubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsystemCgroupPath, err := GetCgroupPath(ms.Name(), cgroupPath, true); err != nil {
		return err
	} else {
		if res.MemoryLimit != "" {
			if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "memory.limit_in_bytes"),
				[]byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf ("set cgroup memory fail %v", err)
			}
		}
		return nil
	}
}

func (ms *MemorySubsystem) Apply(cgroupPath string, pid int) error {
	if subsystemCgroupPath, err := GetCgroupPath(ms.Name(), cgroupPath, false); err != nil {
		return fmt.Errorf ("get cgroup %v error: %v", cgroupPath , err)
	} else {
		if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf ("apply cgroup proc fail %v", err)
		}
		return nil
	}
}

func (ms *MemorySubsystem) Remove(cgroupPath string) error {
	if subsystemCgroupPath, err := GetCgroupPath(ms.Name(), cgroupPath, false); err != nil {
		return fmt.Errorf ("get cgroup %v error: %v", cgroupPath , err)
	} else {
		return os.RemoveAll(subsystemCgroupPath)
	}
}


