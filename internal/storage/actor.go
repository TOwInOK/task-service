package storage

import "fmt"

// Actor — единственный владелец данных задач.
// Все операции проходят через канал команд, конкурентность отсутствует.
type Actor struct {
	cmdChan   chan Command
	storeDir  string
	initTasks map[string]Task // устанавливается в New(), забирается в run()
}

// New создаёт Actor, загружает задачи из хранилища и запускает горутину обработки.
func New(storeDir string) (*Actor, error) {
	if err := ensureStoreDir(storeDir); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}

	tasks, err := loadTasks(storeDir)
	if err != nil {
		return nil, fmt.Errorf("load tasks: %w", err)
	}

	a := &Actor{
		cmdChan:   make(chan Command, 1),
		storeDir:  storeDir,
		initTasks: tasks,
	}

	go a.run()
	return a, nil
}

// Send отправляет команду актору.
func (a *Actor) Send(cmd Command) {
	a.cmdChan <- cmd
}

// Stop завершает работу актора.
func (a *Actor) Stop() {
	close(a.cmdChan)
}

// run — основной цикл обработки команд.
func (a *Actor) run() {
	tasks := a.initTasks
	a.initTasks = nil

	for cmd := range a.cmdChan {
		switch c := cmd.(type) {
		case GetAllCommand:
			a.handleGetAll(c, tasks)

		case GetByIDCommand:
			a.handleGetByID(c, tasks)

		case CreateCommand:
			a.handleCreate(c, tasks)

		case UpdateCommand:
			a.handleUpdate(c, tasks)

		case DeleteCommand:
			a.handleDelete(c, tasks)
		}
	}
}

func (a *Actor) handleGetAll(cmd GetAllCommand, tasks map[string]Task) {
	all := make([]Task, 0, len(tasks))
	for _, t := range tasks {
		all = append(all, t)
	}
	cmd.Result <- Result{Tasks: all}
}

func (a *Actor) handleGetByID(cmd GetByIDCommand, tasks map[string]Task) {
	task, ok := tasks[cmd.ID]
	if !ok {
		cmd.Result <- Result{Err: fmt.Errorf("task %s not found", cmd.ID)}
		return
	}
	cmd.Result <- Result{Task: &task}
}

func (a *Actor) handleCreate(cmd CreateCommand, tasks map[string]Task) {
	task := cmd.Task
	task.ID = generateUUID()
	tasks[task.ID] = task

	if err := writeTaskFile(a.storeDir, task); err != nil {
		delete(tasks, task.ID)
		cmd.Result <- Result{Err: fmt.Errorf("save task: %w", err)}
		return
	}

	_ = writeLockFile(a.storeDir, tasks)
	cmd.Result <- Result{Task: &task}
}

func (a *Actor) handleUpdate(cmd UpdateCommand, tasks map[string]Task) {
	old, ok := tasks[cmd.ID]
	if !ok {
		cmd.Result <- Result{Err: fmt.Errorf("task %s not found", cmd.ID)}
		return
	}

	task := cmd.Task
	task.ID = cmd.ID
	tasks[cmd.ID] = task

	if err := writeTaskFile(a.storeDir, task); err != nil {
		tasks[cmd.ID] = old // откат
		cmd.Result <- Result{Err: fmt.Errorf("save task: %w", err)}
		return
	}

	_ = writeLockFile(a.storeDir, tasks)
	cmd.Result <- Result{Task: &task}
}

func (a *Actor) handleDelete(cmd DeleteCommand, tasks map[string]Task) {
	task, ok := tasks[cmd.ID]
	if !ok {
		cmd.Result <- Result{Err: fmt.Errorf("task %s not found", cmd.ID)}
		return
	}

	delete(tasks, cmd.ID)

	if err := deleteTaskFile(a.storeDir, cmd.ID); err != nil {
		tasks[cmd.ID] = task // откат
		cmd.Result <- Result{Err: fmt.Errorf("delete task file: %w", err)}
		return
	}

	_ = writeLockFile(a.storeDir, tasks)
	cmd.Result <- Result{}
}
