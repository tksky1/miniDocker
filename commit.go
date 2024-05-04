package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

func commitContainer(imageName string) {
	mnt := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mnt, ".").CombinedOutput(); err != nil {
		log.Errorf("commit error: %v", err)
	} else {
		fmt.Println("committed " + imageTar)
	}
}
