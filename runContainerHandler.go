package main

import (
	log "github.com/sirupsen/logrus"
	"miniDocker/persist"
	subsystems2 "miniDocker/subsystems"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// NewParentProcess 准备一个在新namespace运行的新进程
func NewParentProcess(tty bool, volume string) (*exec.Cmd, *os.File) {
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
	mnt := "/root/mnt/"
	root := "/root/"
	newWorkSpace(root, mnt, volume)
	retCmd.Dir = mnt
	return retCmd, writePipe
}

// RunHandler 处理minidocker run,拉起新进程运行minidocker init
func RunHandler(tty bool, cmd []string, res *subsystems2.ResourceConfig, volume string, name string) {

	parentProcess, writePipe := NewParentProcess(tty, volume)
	if parentProcess == nil {
		log.Error("Error creating parent process")
		return
	}

	if err := parentProcess.Start(); err != nil {
		log.Error(err)
	}

	log.Infof("creating container %s", name)
	persist.SaveContainer(&persist.ContainerRecord{
		Pid:        parentProcess.Process.Pid,
		Name:       name,
		Command:    strings.Join(cmd, " "),
		CreateTime: time.Now().Format(time.ANSIC),
		Status:     persist.STATRUNNING,
	})

	// 为容器创建新的cgroup
	cgroup := subsystems2.NewCgroup("minidocker-cgroup")
	defer cgroup.Destroy()
	err := cgroup.Set(res)
	if err != nil {
		log.Error(err)
	}

	if parentProcess.Process == nil {
		log.Fatalf("failed to run container process, is sudo used?")
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

	if tty {
		_ = parentProcess.Wait()
		persist.DeleteRecord(name)
		mnt := "/root/mnt/"
		root := "/root/"
		DeleteWorkSpace(root, mnt, volume)
		log.Infof("容器已退出！")
	}

	os.Exit(0)
}
