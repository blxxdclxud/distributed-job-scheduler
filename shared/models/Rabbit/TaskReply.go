package Rabbit

type TaskReply struct {
	Results  string
	WorkerId string
	JobId    string
	Err      string
}

type TaskReplyWrapper struct {
	TaskReply TaskReply
	Err       error
}
