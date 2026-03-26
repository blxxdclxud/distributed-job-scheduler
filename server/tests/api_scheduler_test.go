package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
	"net/http"
	"testing"
)

type job struct {
	priority models.JobPriority
	script   string
}

func TestAPIAndScheduler(t *testing.T) {

	testData := []job{
		{
			priority: models.LowPriority,
			script:   `local start = os.time(); while os.time() - start < 2 do end; local a, b = 10, 20; return a * b`,
		},
		{
			priority: models.LowPriority,
			script:   `local start = os.time(); while os.time() - start < 3 do end; local a, b = 10, 20; return a * b`,
		},
		{
			priority: models.HighPriority,
			script:   `local start = os.time(); while os.time() - start < 5 do end; local a, b = 10, 20; return a * b`,
		},
	}

	server.RunServer()

	for _, j := range testData {

		buff, _ := json.Marshal(j)
		resp, err := http.Post(
			"http://localhost:8080",
			"application/json",
			bytes.NewReader(buff),
		)
		fmt.Println(resp, err)
	}

}
