package storage

import "fmt"

type Status int

const (
	Draft Status = iota
	InProgress
	Done
)

func (s Status) String() string {
	switch s {
	case Draft:
		return "Draft"
	case InProgress:
		return "In Progress"
	case Done:
		return "Done"
	default:
		return "Unknown"
	}
}

func ParseStatus(s string) (Status, error) {
	switch s {
	case "Draft":
		return Draft, nil
	case "In Progress":
		return InProgress, nil
	case "Done":
		return Done, nil
	default:
		return Draft, fmt.Errorf("unknown status: %s", s)
	}
}
