package storage

// Command — sealed interface.
// Приватный метод isCommand() не даёт реализовать интерфейс вне пакета.
type Command interface {
	isCommand()
}

type GetAllCommand struct {
	Result chan Result
}

type GetByIDCommand struct {
	ID     string
	Result chan Result
}

type CreateCommand struct {
	Task   Task
	Result chan Result
}

type UpdateCommand struct {
	ID     string
	Task   Task
	Result chan Result
}

type DeleteCommand struct {
	ID     string
	Result chan Result
}

func (GetAllCommand) isCommand()  {}
func (GetByIDCommand) isCommand() {}
func (CreateCommand) isCommand()  {}
func (UpdateCommand) isCommand()  {}
func (DeleteCommand) isCommand()  {}
