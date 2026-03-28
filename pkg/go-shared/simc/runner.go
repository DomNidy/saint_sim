package simc

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
)

type Runner struct {
	binaryPath string
}

func NewRunner(binaryPath string) Runner {
	return Runner{binaryPath: binaryPath}
}

func (r Runner) Perform(options string) ([]byte, error) {
	if len(options) == 0 {
		return nil, errors.New("options cannot be empty")
	}

	// Command to invoke simc and perform the sim
	simCommand := exec.Command(r.binaryPath, options)

	// Capture output of sim command and write it to this buffer
	var outputBuffer bytes.Buffer
	simCommand.Stdout = &outputBuffer
	simCommand.Stderr = os.Stderr

	// Run the sim command
	if err := simCommand.Run(); err != nil {
		log.Printf("Failed to execute sim binary: %v", err)
		return nil, err
	}

	// Get the output as a byte slice
	simResult := outputBuffer.Bytes()

	return simResult, nil
}

func (r Runner) Version() string {
	log.Print(r.binaryPath)
	simcCommand := exec.Command(r.binaryPath)

	var outputBuffer bytes.Buffer
	simcCommand.Stdout = &outputBuffer

	err := simcCommand.Run()

	exitCode := simcCommand.ProcessState.ExitCode()
	const noArgumentsExitCode = 50 // simc returns exitcode 50 when no arguments are provided.
	// since we just want to read its stdout to parse version number, we can ignore the err

	if err != nil && exitCode != noArgumentsExitCode {
		log.Fatalf("Error running simc binary: %v", err)
	}

	return outputBuffer.String()
}
