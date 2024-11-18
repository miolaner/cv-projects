package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Create a simulated map file content with more than 10,000 stations
	var mapContent strings.Builder
	mapContent.WriteString("stations:\n")
	for i := 0; i < 10001; i++ {
		mapContent.WriteString(fmt.Sprintf("station_%d,%d,%d\n", i, i, i))
	}
	mapContent.WriteString("connections:\n")
	for i := 0; i < 10000; i++ {
		mapContent.WriteString(fmt.Sprintf("station_%d-station_%d\n", i, i+1))
	}

	// Write the simulated map content to a temporary file
	tmpFile, err := os.CreateTemp("", "map_*.txt")
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(mapContent.String())
	if err != nil {
		fmt.Println("Error writing to temporary file:", err)
		return
	}
	tmpFile.Close()

	// Prepare the command to run the main program
	cmd := exec.Command("go", "run", "main.go", tmpFile.Name(), "station_0", "station_10000", "1")

	// Capture stderr
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run the command
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error executing command:", err)
	}

	// Check and print the captured stderr
	output := stderr.String()
	fmt.Println("Captured stderr:", output)

	// Verify the error message
	if strings.Contains(output, "Error: Train map exceeded the maximum number(10,000) of allowed stations, exiting...") {
		fmt.Println("Test passed: Correct error message captured")
	} else {
		fmt.Println("Test failed: Incorrect error message")
	}
}
