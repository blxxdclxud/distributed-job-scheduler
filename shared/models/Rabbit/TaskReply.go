package Rabbit

type TaskReply struct {
	Results  interface{}
	WorkerId string
	JobId    string
	Err      error
}
