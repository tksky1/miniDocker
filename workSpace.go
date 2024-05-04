package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

// 为容器挂载好工作空间
func newWorkSpace(root string, mnt string, volume string) {
	// 删除原先绑定的环境
	DeleteWorkSpace(root, mnt, volume)

	createReadOnlyLayer(root)
	createWriteLayer(root)
	createMountPoint(root, mnt)

	// 挂载数据卷，如果指定的话
	if volume != "" {
		volumeUrls := volumeUrlExtract(volume)
		if len(volumeUrls) == 2 && volumeUrls[0] != "" && volumeUrls[1] != "" {
			MountVolume(mnt, volumeUrls)
		} else {
			log.Info("Invalid volume input!")
		}
	}
}

// MountVolume 挂载数据卷
func MountVolume(mnt string, volumeUrls []string) {
	parentUrl := volumeUrls[0]
	if err := os.Mkdir(parentUrl, 0777); err != nil {
		log.Infof("mkdir for parent dir %s failed :%v", parentUrl, err)
	}
	// 在容器文件系统内创建挂载点
	containerUrl := volumeUrls[1]
	mntContainer := mnt + containerUrl
	if err := os.Mkdir(mntContainer, 0777); err != nil {
		log.Infof("mkdir for in-container dir %s failed :%v", mntContainer, err)
	}
	// 把宿主机目录挂载到容器内挂载点
	dirs := "dirs=" + parentUrl
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntContainer)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount volume fail: %v", err)
	}
}

func createReadOnlyLayer(root string) {
	busybox := root + "busybox/"
	exist, err := pathExists(busybox)
	if err != nil || exist == false {
		if _, err := exec.Command("mkdir", "/root/busybox").CombinedOutput(); err != nil {
			log.Errorf("mkdir error: %v", err)
		}
		if _, err := exec.Command("tar", "-xvf", "busybox.tar", "-C", "/root/busybox").CombinedOutput(); err != nil {
			log.Errorf("install busybox error: %v", err)
		} else {
			log.Info("installed busybox")
			return
		}
		log.Errorf("fail to find busybox path:%v", err)
	}
}

func createWriteLayer(root string) {
	writeLayer := root + "writeLayer/"
	if err := os.Mkdir(writeLayer, 0777); err != nil {
		log.Error(err)
	}
}

// 挂载到新建的mnt目录
func createMountPoint(root string, mnt string) {
	// 创建mnt
	if err := os.Mkdir(mnt, 0777); err != nil {
		log.Errorf("mkdir err: %v", err)
	}
	// 使用aufs将只读和可写目录都mount到mnt目录下
	dirs := "dirs=" + root + "writeLayer:" + root + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mnt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("run mount fail: %v", err)
	}
}

func DeleteWorkSpace(root string, mnt string, volume string) {
	if volume != "" {
		volumeUrls := volumeUrlExtract(volume)
		if len(volumeUrls) == 2 && volumeUrls[0] != "" && volumeUrls[1] != "" {
			DeleteMountPoint(mnt+volumeUrls[1], true)
		}
	}
	DeleteMountPoint(mnt, true)
	DeleteWriteLayer(root)
}

func DeleteMountPoint(mnt string, remove bool) {
	cmd := exec.Command("umount", mnt)
	cmd.Run()
	if !remove {
		return
	}
	if err := os.RemoveAll(mnt); err != nil {
		log.Errorf("remove dir fail: %v", err)
	}
}

func DeleteWriteLayer(root string) {
	write := root + "writeLayer"
	if err := os.RemoveAll(write); err != nil {
		log.Error(err)
	}
}

// 工具函数，提取挂载前后目录名
func volumeUrlExtract(volume string) []string {
	var volumeUrls []string
	volumeUrls = strings.Split(volume, ":")
	return volumeUrls
}

// 工具函数，检测路径是否存在
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
