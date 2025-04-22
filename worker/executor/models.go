package executor

type TaskReply struct {
	Results  interface{}
	WorkerId string
	Err      error
}
