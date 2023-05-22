package subsystems

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"strings"
)

// ResourceConfig 表征资源限制
type ResourceConfig struct {
	MemoryLimit string
	CPUShare    string
	CPUSet      string
}

// Subsystem 接口
type Subsystem interface {
	// Name 返回subsystem的名字
	Name() string
	// Set 设置某cgroup在该subsystem中的资源限制
	Set(path string, res *ResourceConfig) error
	// Apply 将进程添加到某cgroup中
	Apply(path string, pid int) error
	// Remove 移除cgroup
	Remove(path string) error
}

var SubsystemIns = []Subsystem{
	&CPUSubsystem{},
	&CPUSetSubsystem{},
	&MemorySubsystem{},
}

// FindCgroupMountPoint 从/proc/self/mountinfo 查找挂载该subsystem的hierarchy的cgroup目录
func FindCgroupMountPoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		log.Error(err)
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4]
				// 把对应的段取出来
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Error(err)
		return ""
	}
	return ""
}

// GetCgroupPath 取Cgroup在文件系统内的绝对路径
func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountPoint(subsystem)
	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err == nil ||
		(autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err != nil {
				return "", fmt.Errorf("cannot create cgroup: %v", err)
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	} else {
		return "", fmt.Errorf("cgroup path error: %v", err)
	}
}
