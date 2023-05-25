package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// RunContainerInitProcess 初始化新容器
func RunContainerInitProcess() error {

	// 读index为3的文件描述符，即之前传过来的管道
	pipe := os.NewFile(uintptr(3), "pipe")
	// 读父进程传过来的内容，可能会阻塞
	msg, err := io.ReadAll(pipe)
	if err != nil {
		log.Error("init read pipe error %v", err)
		return nil
	}
	str := string(msg)
	cmd := strings.Split(str, " ")

	log.Infof("Initiating %s", str)

	if cmd == nil || len(cmd) == 0 {
		return fmt.Errorf("run container fail: cmd is nil")
	}

	setupMount()

	path, err := exec.LookPath(cmd[0])
	if err != nil {
		log.Error(err)
	}

	// 启动新进程取代现有进程，执行init命令
	err = syscall.Exec(path, cmd[:], os.Environ())
	if err != nil {
		log.Error(err)
	}
	return nil
}

// 初始化挂载点
func setupMount() {

	// 声明命名空间是独立的，避免再次运行出现bug
	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		log.Error(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Error(err)
		return
	}
	err = pivotRoot(pwd)
	if err != nil {
		log.Errorf("pivot failed: %v", err)
		return
	}

	// 为proc文件系统设置mount参数
	mountFlags := syscall.MS_NOEXEC | // 不允许其他进程访问
		syscall.MS_NOSUID | // 禁止其他进程 set-user-ID / set-group-ID
		syscall.MS_NODEV
	// 挂载proc filesystem
	err = syscall.Mount("proc", "/proc", "proc", uintptr(mountFlags), "")
	if err != nil {
		log.Error(err)
	}

	// 为dev挂载一个基于内存的临时文件系统
	err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME,
		"mode=755")
	if err != nil {
		log.Error(err)
	}

}

// 使用系统调用pivot_root来给当前进程创建一个新的root文件系统
func pivotRoot(root string) error {
	// 重新挂载一次，从而为root创建一个新的fs
	err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, "")
	if err != nil {
		return fmt.Errorf("fail remount root:%v", err)
	}
	// 旧root放在这（rootfs/.pivot_root)
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	// 将当前进程的root fs移动到pivotDir，并使"root"成为新的文件系统
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot fail: %v", err)
	}

	// 修改当前工作目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir fail: %v", err)
	}

	// 删除老root
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("fail unmount pivotdir")
	}
	return os.Remove(pivotDir)
}
