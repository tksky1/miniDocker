package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"miniDocker/main/subsystems"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// NewParentProcess 准备一个在新namespace运行的新进程
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	readPipe, writePipe, _ := os.Pipe()
	retCmd := exec.Command("/proc/self/exe", "init")
	// 设置新namespace的参数以完成隔离
	retCmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		retCmd.Stdin = os.Stdin
		retCmd.Stdout = os.Stdout
		retCmd.Stderr = os.Stderr
	}
	// 传入管道文件读取端的句柄
	retCmd.ExtraFiles = []*os.File{readPipe}
	return retCmd, writePipe
}

// RunHandler 处理minidocker run,拉起新进程运行minidocker init
func RunHandler(tty bool, cmd []string, res *subsystems.ResourceConfig) {

	parentProcess, writePipe := NewParentProcess(tty)
	if parentProcess == nil {
		log.Error("Error creating parent process")
		return
	}

	if err := parentProcess.Start(); err != nil {
		log.Error(err)
	}

	// 为容器创建新的cgroup
	cgroup := subsystems.NewCgroup("minidocker-cgroup")
	defer cgroup.Destroy()
	err := cgroup.Set(res)
	if err != nil {
		log.Error(err)
	}

	if err = cgroup.Apply(parentProcess.Process.Pid); err != nil {
		log.Error(err)
	}

	// 传送指令到父进程
	commands := strings.Join(cmd, " ")
	_, err = writePipe.WriteString(commands)
	writePipe.Close()
	if err != nil {
		return
	}

	_ = parentProcess.Wait()
	os.Exit(0)
}

// RunContainerInitProcess 初始化新容器
func RunContainerInitProcess() error {

	// 声明命名空间是独立的，避免再次运行出现bug
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")

	// 为proc文件系统设置mount参数
	mountFlags := syscall.MS_NOEXEC | // 不允许其他进程访问
		syscall.MS_NOSUID | // 禁止其他进程 set-user-ID / set-group-ID
		syscall.MS_NODEV
	// 挂载proc filesystem
	err := syscall.Mount("proc", "/proc", "proc", uintptr(mountFlags), "")
	if err != nil {
		log.Error(err)
	}

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
