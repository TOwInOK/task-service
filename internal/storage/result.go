package storage

type Result struct {
	Task  *Task
	Tasks []Task
	Err   error
}
