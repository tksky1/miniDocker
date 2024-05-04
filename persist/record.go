package persist

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path"
)

const (
	STATRUNNING     = "running"
	STATSTOPPED     = "stopped"
	STATEXIT        = "exited"
	PERSISTLOCATION = "/var/run/minidocker"
	PERSISTNAME     = "info.yaml"
)

type ContainerRecord struct {
	Pid        int    `yaml:"pid"`
	Name       string `yaml:"name"`
	Command    string `yaml:"command"`
	CreateTime string `yaml:"createTime"`
	Status     string `yaml:"status"`
}

// SaveContainer 把容器信息持久化为 yaml
func SaveContainer(r *ContainerRecord) {
	testMiniDockerDir()
	savePath := path.Join(PERSISTLOCATION, r.Name)

	if _, err := os.Stat(savePath); err == nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(savePath, 0755); err != nil {
				log.Errorf("error mkdir: %v", err)
			}
		}
	}

	saveFilePath := path.Join(savePath, PERSISTNAME)

	if _, err := os.Stat(savePath); err == nil || !os.IsNotExist(err) {
		err := os.RemoveAll(savePath)
		if err != nil {
			log.Errorf("removing dir error: %v", err)
		}
	}

	if err := os.Mkdir(savePath, 0755); err != nil {
		log.Errorf("error mkdir: %v", err)
	}

	file, err := os.Create(saveFilePath)
	if err != nil {
		log.Errorf("error creating file:%v", err)
		return
	}
	defer file.Close()
	encoder := yaml.NewEncoder(file)

	err = encoder.Encode(r)
	if err != nil {
		log.Errorf("error encoding YAML: %v", err)
		return
	}
}

func DeleteRecord(name string) {
	if !testMiniDockerDir() {
		return
	}
	record := path.Join(PERSISTLOCATION, name)
	if err := os.RemoveAll(record); err != nil {
		log.Errorf("delete container error: %v", err)
	}
}

func GetRecord(name string) *ContainerRecord {
	testMiniDockerDir()
	file, err := os.Open(path.Join(PERSISTLOCATION, name, PERSISTNAME))
	if err != nil {
		log.Errorf("error loading container info for %s : %v", name, err)
		return nil
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	var info ContainerRecord
	err = decoder.Decode(&info)
	if err != nil {
		log.Errorf("error parsing container info for %s : %v", name, err)
		return nil
	}
	return &info
}

func testMiniDockerDir() bool {
	if _, err := os.Stat(PERSISTLOCATION); err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(PERSISTLOCATION, 0755); err != nil {
			log.Errorf("error mkdir: %v", err)
		}
		return false
	}
	return true
}
