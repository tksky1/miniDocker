package subsystems

import log "github.com/sirupsen/logrus"

type Cgroup struct {
	// cgroup在hierarchy中的路径，是相对路径
	Path string
	// 资源限制
	Resource *ResourceConfig
}

func NewCgroup(path string) *Cgroup {
	return &Cgroup{Path: path}
}

// Apply 将线程PID加入到每个subsystem挂载的Cgroup中
func (c *Cgroup) Apply(pid int) error {
	for _, subSysIns := range SubsystemIns {
		err := subSysIns.Apply(c.Path, pid)
		if err != nil {
			log.Error(err)
		}
	}
	return nil
}

// Set 设置每个subsystem挂载的cgroup的资源限制
func (c *Cgroup) Set(res *ResourceConfig) error {
	for _, subSysIns := range SubsystemIns {
		err := subSysIns.Set(c.Path, res)
		if err != nil {
			log.Error(err)
		}
	}
	return nil
}

// Destroy 释放各subsystem挂载的cgroup
func (c *Cgroup) Destroy() error {
	for _, subsys := range SubsystemIns {
		if err := subsys.Remove(c.Path); err != nil {
			log.Warnf("remove cgroup fail: %v", err)
		}
	}
	return nil
}
