package domain

type Id uint32
type TaskStatus string

const (
	StatusNew  TaskStatus = "NEW"
	StatusDone TaskStatus = "DONE"
)

type TaskCreateRequest struct {
	Title string
	Text  string
}

type Task struct {
	ID     Id
	Title  string
	Text   string
	Status TaskStatus
}
