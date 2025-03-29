package storage

import (
	"dns-monitor/internal/common"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func LoadPreviousState() (common.PreviousState, error) {
	stateFile, err := getStateFileLocation()
	if err != nil {
		log.Printf("Error retrieving the state file path: %v", err)
		return common.PreviousState{}, fmt.Errorf("failed to retrieve the state file path: %w", err)
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Return an empty state with no error
			log.Println("ℹ️ State file does not exist, returning empty state. ℹ️")
			return common.PreviousState{}, nil
		}
		log.Printf("Error reading state file: %v", err)
		return common.PreviousState{}, err
	}

	var state common.PreviousState
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("Error parsing state file: %v", err)
		return common.PreviousState{}, err
	}

	return state, nil
}

func SavePreviousState(state common.PreviousState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Printf("Error serializing state: %v", err)
		return fmt.Errorf("failed to serialize state: %w", err)
	}

	stateFile, err := getStateFileLocation()
	if err != nil {
		log.Printf("Error retrieving the state file path: %v", err)
		return fmt.Errorf("failed to retrieve the state file path: %w", err)
	}

	err = os.WriteFile(stateFile, data, 0644)
	if err != nil {
		log.Printf("Error writing state file: %v", err)
		return fmt.Errorf("failed to write state file: %w", err)
	}

	log.Printf("✅ successfully saved the state ✅")

	return nil
}

func getStateFileLocation() (string, error) {
	const fileName = "/dns_state.json"
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}

	dir := filepath.Join(workingDir, "data")

	// Create the directory (including parents if necessary)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Error creating directory: %v", err)
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	return dir + fileName, nil
}
