package Rabbit

type HealthReport struct {
	WorkerId  string `json:"workerId"`
	TimeStamp int64  `json:"timestamp"`
}
