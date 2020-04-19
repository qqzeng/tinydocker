package cgroups

//import "./subsystems"
import "github.com/qqzeng/tinydocker/cgroups/subsystems"

type CgroupManager struct {
	Path string
	Resouce *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{Path:path}
}

func (cm *CgroupManager) Apply(pid int) error {
	for _, subsystemIns := range(subsystems.SubsystemInstances) {
		if err := subsystemIns.Apply(cm.Path, pid); err != nil {
			return err
		}
	}
	return nil
}

func (cm *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subsystemIns := range(subsystems.SubsystemInstances) {
		if err := subsystemIns.Set(cm.Path, res); err != nil {
			return err
		}
	}
	return nil
}

func (cm *CgroupManager) Destory() error {
	for _, subsystemIns := range(subsystems.SubsystemInstances) {
		if err := subsystemIns.Remove(cm.Path); err != nil {
			return err
		}
	}
	return nil
}