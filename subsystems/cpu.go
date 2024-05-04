package subsystems

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"strconv"
)

// CPUSubsystem memory的subsystem，实现subsystem接口
type CPUSubsystem struct {
}

func (s *CPUSubsystem) Name() string {
	return "cpu"
}

// Set 设置某Cgroup的内存限制
func (s *CPUSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	// 取得当前subsystem在虚拟文件系统的路径
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.CPUShare != "" {
			if err := os.WriteFile(path.Join(subsysCgroupPath, "cpu.shares"),
				[]byte(res.CPUShare), 0644); err != nil {
				return fmt.Errorf("set cgroup cpu fail: %v", err)
			}
			log.Info("cpu share set: %s", res.CPUShare)
		}
		return nil
	} else {
		return err
	}
}

// Remove 删除对应cgroup
func (s *CPUSubsystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.Remove(subsysCgroupPath)
	} else {
		return err
	}
}

// Apply 添加一个进程到该cgroup中
func (s *CPUSubsystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := os.WriteFile(path.Join(subsysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("apply cpu cgroup process fail: %v", err)
		}
		log.Infof("apply pid %d succeed to %s", pid, subsysCgroupPath)
		return nil
	} else {
		return err
	}
}
