package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"task-service/internal/storage"
)

type Handler struct {
	actor *storage.Actor
}

func NewHandler(actor *storage.Actor) *Handler {
	return &Handler{actor: actor}
}

// sendCommand sends a command to the actor and waits for the result.
func (h *Handler) sendCommand(cmd storage.Command) storage.Result {
	h.actor.Send(cmd)
	switch c := cmd.(type) {
	case storage.GetAllCommand:
		return <-c.Result
	case storage.GetByIDCommand:
		return <-c.Result
	case storage.CreateCommand:
		return <-c.Result
	case storage.UpdateCommand:
		return <-c.Result
	case storage.DeleteCommand:
		return <-c.Result
	default:
		return storage.Result{Err: fmt.Errorf("unknown command type")}
	}
}

// ---------- Pages ----------

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	tasks := h.getAllTasks()
	draftCount := 0
	for _, t := range tasks {
		if t.Status == storage.Draft {
			draftCount++
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(renderIndex(tasks, draftCount)))
}

// ---------- HTMX Partials ----------

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cmd := storage.CreateCommand{
		Task:   storage.Task{Title: title, Status: storage.Draft, Description: r.FormValue("description")},
		Result: make(chan storage.Result, 1),
	}
	res := h.sendCommand(cmd)
	if res.Err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tasks := h.getAllTasks()
	draftCount := 0
	for _, t := range tasks {
		if t.Status == storage.Draft {
			draftCount++
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(renderMainArea(tasks, draftCount)))
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	statusStr := r.FormValue("status")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status, err := storage.ParseStatus(statusStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	getCmd := storage.GetByIDCommand{ID: id, Result: make(chan storage.Result, 1)}
	res := h.sendCommand(getCmd)
	if res.Err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	task := *res.Task
	task.TransitionTo(status)

	updateCmd := storage.UpdateCommand{ID: id, Task: task, Result: make(chan storage.Result, 1)}
	res = h.sendCommand(updateCmd)
	if res.Err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tasks := h.getAllTasks()
	draftCount := 0
	for _, t := range tasks {
		if t.Status == storage.Draft {
			draftCount++
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(renderMainArea(tasks, draftCount)))
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cmd := storage.DeleteCommand{ID: id, Result: make(chan storage.Result, 1)}
	res := h.sendCommand(cmd)
	if res.Err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tasks := h.getAllTasks()
	draftCount := 0
	for _, t := range tasks {
		if t.Status == storage.Draft {
			draftCount++
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(renderMainArea(tasks, draftCount)))
}

// ---------- API JSON ----------

func (h *Handler) APICreateTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	cmd := storage.CreateCommand{
		Task:   storage.Task{Title: input.Title, Description: input.Description, Status: storage.Draft},
		Result: make(chan storage.Result, 1),
	}
	res := h.sendCommand(cmd)
	if res.Err != nil {
		http.Error(w, res.Err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Task)
}

func (h *Handler) APIGetTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.getAllTasks()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) APIGetTask(w http.ResponseWriter, r *http.Request) {
	id := chiURLParam(r, "id")
	cmd := storage.GetByIDCommand{ID: id, Result: make(chan storage.Result, 1)}
	res := h.sendCommand(cmd)
	if res.Err != nil {
		http.Error(w, res.Err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Task)
}

func (h *Handler) APIUpdateTask(w http.ResponseWriter, r *http.Request) {
	id := chiURLParam(r, "id")
	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	status, err := storage.ParseStatus(input.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch existing task to preserve time tracking fields
	getCmd := storage.GetByIDCommand{ID: id, Result: make(chan storage.Result, 1)}
	res := h.sendCommand(getCmd)
	if res.Err != nil {
		http.Error(w, res.Err.Error(), http.StatusNotFound)
		return
	}

	task := *res.Task
	task.Title = input.Title
	task.Description = input.Description
	task.TransitionTo(status)

	updateCmd := storage.UpdateCommand{ID: id, Task: task, Result: make(chan storage.Result, 1)}
	res = h.sendCommand(updateCmd)
	if res.Err != nil {
		http.Error(w, res.Err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Task)
}

func (h *Handler) APIDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := chiURLParam(r, "id")
	cmd := storage.DeleteCommand{ID: id, Result: make(chan storage.Result, 1)}
	res := h.sendCommand(cmd)
	if res.Err != nil {
		http.Error(w, res.Err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- Helpers ----------

func (h *Handler) getAllTasks() []storage.Task {
	cmd := storage.GetAllCommand{Result: make(chan storage.Result, 1)}
	res := h.sendCommand(cmd)
	if res.Err != nil {
		return nil
	}
	return res.Tasks
}
