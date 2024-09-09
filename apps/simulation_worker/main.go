package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/DomNidy/saint_sim/pkg/secrets"
)

func performSim(region, realm, name string) {
	simcBinaryPath := secrets.LoadSecret("SIMC_BINARY_PATH")
	outputFilePath := fmt.Sprintf("output=%v-%v-%v-%d.txt", region, realm, name, time.Now().Unix())
	simCommand := exec.Cmd{
		Path:   simcBinaryPath.Value(),
		Args:   []string{simcBinaryPath.Value(), fmt.Sprintf("armory=%v,%v,%v", region, realm, name), outputFilePath},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	fmt.Println(simCommand.Args)
	if err := simCommand.Run(); err != nil {
		log.Fatalf("Failed to execute sim binary: %v", err)
	}

}

// todo: consume messages from rabbitmq queue
// todo: write the results of the sim to database
func main() {
	performSim("us", "hydraxis", "ishton")
}
