package Rabbit

type TaskReply struct {
	Results  interface{}
	WorkerId string
	Err      error
}
