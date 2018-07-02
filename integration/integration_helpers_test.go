package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/containous/traefik/log"
)

func startTestCluster() {
	fmt.Println("INTEGRATION TEST: Integrations tests run against a 5 node cluster running inside a docker image.")
	fmt.Println("INTEGRATION TEST: Depending on machine spec these can take ~5mins to run end to end.")
	fmt.Println("INTEGRATION TEST: Use `sfintegration.verbose` flag to show full output")
	fmt.Println("INTEGRATION TEST: Starting cluster....")

	_, err := runScript("run.sh", time.Second*900)
	if err != nil {
		panic(err)
	}

	fmt.Println("INTEGRATION TEST: Cluster started successfully.")
}

func stopTestCluster() {
	_, err := runScript("stop.sh", time.Second*30)
	if err != nil {
		panic("Failed to stop cluster")
	}
}

// func resetTestCluster(t *testing.T) string {
// 	output, err := runScript("reset.sh", time.Second*80)
// 	if err != nil {
// 		t.Error(err)
// 		t.Log(output)
// 		t.FailNow()
// 	}

// 	return output
// }

func runScript(scriptName string, timeout time.Duration) (string, error) {
	resultChan := make(chan string, 1)
	failedChan := make(chan error, 1)

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("/bin/sh", filepath.Join(dir, "scripts", scriptName))

	go func() {
		cmd.Dir = filepath.Join(dir, "scripts")

		if isVerbose {
			fmt.Println("INTEGRATION TEST: Using verbose script output")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout

			err := cmd.Run()
			if err != nil {
				failedChan <- err
				return
			}
			resultChan <- ""
		} else {
			output, err := cmd.CombinedOutput()
			resultChan <- string(output)
			if err != nil {
				log.Infof("Failed running script: %v", err)
			}
		}

	}()

	select {
	case err := <-failedChan:
		return "", err
	case output := <-resultChan:
		return string(output), nil
	case <-time.After(timeout):
		cmd.Process.Kill()
		return "", fmt.Errorf("Timeout waiting for script after: %v", timeout)
	}
}

func toJSON(i interface{}) string {
	jsonBytes, err := json.Marshal(i)
	if err != nil {
		panic("Failed to marshal json")
	}

	return string(jsonBytes)
}