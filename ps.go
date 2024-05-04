package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"miniDocker/persist"
	"os"
	"text/tabwriter"
)

func listContainers() {
	files, err := os.ReadDir(persist.PERSISTLOCATION)
	if err != nil {
		log.Fatalf("read containers fail: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprintf(w, "Name\tPID\tStatus\tCommand\tCreateTime\n")
	for _, file := range files {
		r := persist.GetRecord(file.Name())
		if r != nil {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n", r.Name, r.Pid, r.Status, r.Command, r.CreateTime)
		}
	}
	w.Flush()
}
