package main

import (
	"context"
	"testing"

	api "github.com/DomNidy/saint_sim/internal/api"
)

func TestSimulationInputFromOptions(t *testing.T) {
	t.Parallel()

	options := api.SimulationOptions{
		SimcAddonExport: "mage=\"Example\"\nlevel=80",
	}

	input, err := simulationInputFromOptions(options)
	if err != nil {
		t.Fatalf("simulationInputFromOptions() error = %v", err)
	}

	if input.simcAddonExport != options.SimcAddonExport {
		t.Fatalf("simcAddonExport = %q, want %q", input.simcAddonExport, options.SimcAddonExport)
	}
}

func TestSimulationInputFromOptionsRejectsMissingExport(t *testing.T) {
	t.Parallel()

	_, err := simulationInputFromOptions(api.SimulationOptions{
		SimcAddonExport: "",
	})
	if err == nil {
		t.Fatal("simulationInputFromOptions() error = nil, want error")
	}
}

func TestSimcRunnerRunUsesProfileFile(t *testing.T) {
	t.Parallel()

	runner := simcRunner{binaryPath: "/bin/cat"}
	input := simulationInput{
		simcAddonExport: "warrior=\"Example\"\nlevel=80\nspec=arms\n",
	}

	output, err := runner.Run(context.Background(), input)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if string(output) != input.simcAddonExport {
		t.Fatalf("Run() output = %q, want %q", string(output), input.simcAddonExport)
	}
}
