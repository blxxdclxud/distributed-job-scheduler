package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	// Create a Lua script that calculates 5*10
	luaScript := "local a, b = 5, 10; return a * b"

	// Create the job request
	jobRequest := JobRequest{
		Script:   luaScript,
		Priority: 1, // Mid priority
	}

	// Convert request to JSON
	requestBody, err := json.Marshal(jobRequest)
	if err != nil {
		log.Fatalf("Error creating request JSON: %v", err)
	}

	// Submit the job
	fmt.Println("Submitting job to server...")
	resp, err := http.Post("http://localhost:8080/submit_job",
		"application/json", bytes.NewBuffer(requestBody))
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

		statusURL := fmt.Sprintf("http://localhost:8080/status/%d", jobID)
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
