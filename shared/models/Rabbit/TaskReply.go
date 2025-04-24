package Rabbit

type TaskReply struct {
	Results  interface{}
	WorkerId string
	JobId    string
	Err      error
}

type TaskReplyWrapper struct {
	TaskReply TaskReply
	Err       error
}
