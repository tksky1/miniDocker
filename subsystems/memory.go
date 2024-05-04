package subsystems

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"strconv"
)

// MemorySubsystem memory的subsystem，实现subsystem接口
type MemorySubsystem struct {
}

func (s *MemorySubsystem) Name() string {
	return "memory"
}

// Set 设置某Cgroup的内存限制
func (s *MemorySubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	// 取得当前subsystem在虚拟文件系统的路径
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.MemoryLimit != "" {
			if err := os.WriteFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"),
				[]byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set cgroup memory fail: %v", err)
			}
			log.Infof("set: memory at %s set to %s", subsysCgroupPath, res.MemoryLimit)
		}
		return nil
	} else {
		return err
	}
}

// Remove 删除对应cgroup
func (s *MemorySubsystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.Remove(subsysCgroupPath)
	} else {
		return err
	}
}

// Apply 添加一个进程到该cgroup中
func (s *MemorySubsystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := os.WriteFile(path.Join(subsysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("apply memory cgroup process fail: %v", err)

		}
		log.Infof("apply pid %d succeed to %s", pid, subsysCgroupPath)
		return nil
	} else {
		return err
	}
}
