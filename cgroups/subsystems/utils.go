package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

func FindCgroupMountPoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	// 30  27  0:24  I  /sys/fs/cgroup/rnernory  rw , nosuid, nodev , noexec , relatirne  shared : l3  cgroup  cgroup  rw , rnernory
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, " ")
		for _, field := range(strings.Split(fields[len(fields)-1], ",")) {
			if field == subsystem {
				return fields[4]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return ""
	}
	return ""
}

func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountPoint(subsystem)
	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err2 := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err2 != nil {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	} else {
		return "", fmt.Errorf("error create cgroup %v", err)
	}
}


