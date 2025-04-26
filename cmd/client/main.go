package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// JobRequest matches the server's expected request format
type JobRequest struct {
	Script   string `json:"script"`
	Priority int    `json:"priority,omitempty"`
}

// JobResponse matches the server's response format
type JobResponse struct {
	JobID     int    `json:"job_id"`
	JobStatus string `json:"status"`
	JobResult string `json:"result,omitempty"`
}

func main() {
	// Define command-line flags
	scriptFile := flag.String("file", "-", "File containing the Lua script (use '-' for stdin)")
	serverAddr := flag.String("host", "localhost:8080", "Address of the job server")
	priority := flag.Int("priority", 1, "Job priority (0=high, 1=medium, 2=low)")
	flag.Parse()

	// Read the script from the specified file or stdin
	var script string
	var err error
	if *scriptFile == "-" {
		fmt.Println("Reading Lua script from stdin (type your script and press Ctrl+D when done):")
		scriptBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Error reading from stdin: %v", err)
		}
		script = string(scriptBytes)
	} else {
		scriptBytes, err := os.ReadFile(*scriptFile)
		if err != nil {
			log.Fatalf("Error reading script file %s: %v", *scriptFile, err)
		}
		script = string(scriptBytes)
	}

	if script == "" {
		log.Fatal("No script provided. Please provide a script via file or stdin.")
	}

	// Create the job request
	jobRequest := JobRequest{
		Script:   script,
		Priority: *priority,
	}

	// Convert request to JSON
	requestBody, err := json.Marshal(jobRequest)
	if err != nil {
		log.Fatalf("Error creating request JSON: %v", err)
	}

	// Build the server URL
	serverURL := fmt.Sprintf("http://%s", *serverAddr)
	submitURL := fmt.Sprintf("%s/submit_job", serverURL)

	// Submit the job
	fmt.Printf("Submitting job to server at %s...\n", serverURL)
	resp, err := http.Post(submitURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Error submitting job: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var jobResp JobResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobResp); err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	fmt.Printf("Job submitted successfully! Job ID: %d, Initial status: %s\n",
		jobResp.JobID, jobResp.JobStatus)

	// Poll for the result
	fmt.Println("Polling for job result...")
	jobID := jobResp.JobID

	for {
		time.Sleep(1 * time.Second) // Poll every second

		statusURL := fmt.Sprintf("%s/status/%d", serverURL, jobID)
		statusResp, err := http.Get(statusURL)
		if err != nil {
			log.Printf("Error checking job status: %v", err)
			continue
		}

		var status JobResponse
		if err := json.NewDecoder(statusResp.Body).Decode(&status); err != nil {
			log.Printf("Error parsing status response: %v", err)
			statusResp.Body.Close()
			continue
		}
		statusResp.Body.Close()

		fmt.Printf("Current status: %s\n", status.JobStatus)

		// Check if job is completed or failed
		if status.JobStatus == "COMPLETED" {
			fmt.Printf("Job completed! Result: %s\n", status.JobResult)
			break
		} else if status.JobStatus == "FAILED" {
			fmt.Printf("Job failed: %s\n", status.JobResult)
			break
		}
	}
}
