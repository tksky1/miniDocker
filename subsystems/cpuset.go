package subsystems

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"strconv"
)

// CPUSetSubsystem memory的subsystem，实现subsystem接口
type CPUSetSubsystem struct {
}

func (s *CPUSetSubsystem) Name() string {
	return "cpuset"
}

// Set 设置某Cgroup的内存限制
func (s *CPUSetSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	// 取得当前subsystem在虚拟文件系统的路径
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.CPUSet != "" {
			if err := os.WriteFile(path.Join(subsysCgroupPath, "cpuset.cpus"),
				[]byte(res.CPUSet), 0644); err != nil {
				return fmt.Errorf("set cgroup cpuset fail: %v", err)
			}
			log.Info("cpuset set: %s", res.CPUSet)
		}
		return nil
	} else {
		return err
	}
}

// Remove 删除对应cgroup
func (s *CPUSetSubsystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.Remove(subsysCgroupPath)
	} else {
		return err
	}
}

// Apply 添加一个进程到该cgroup中
func (s *CPUSetSubsystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := os.WriteFile(path.Join(subsysCgroupPath, "cpuset.cpus"),
			[]byte("0"), 0644); err != nil {
			return fmt.Errorf("apply cpuset-cpus cgroup process fail: %v", err)
		}
		if err := os.WriteFile(path.Join(subsysCgroupPath, "cpuset.mems"),
			[]byte("0"), 0644); err != nil {
			return fmt.Errorf("apply cpuset-mems cgroup process fail: %v", err)
		}
		if err := os.WriteFile(path.Join(subsysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("apply cpuset cgroup process fail: %v", err)
		}
		log.Infof("apply pid %d succeed to %s", pid, subsysCgroupPath)
		return nil
	} else {
		return err
	}
}
