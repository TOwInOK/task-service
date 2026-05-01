package storage

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const lockFileName = "lock.json"

// generateUUID генерирует UUID v4.
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ensureStoreDir создаёт директорию хранилища, если её нет.
func ensureStoreDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// loadTasks загружает задачи из lock.json (быстро),
// при ошибке — парсит .md файлы и пересоздаёт lock.json.
func loadTasks(dir string) (map[string]Task, error) {
	tasks, err := loadFromLock(dir)
	if err == nil {
		return tasks, nil
	}

	tasks, err = loadFromFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("load from files: %w", err)
	}

	_ = writeLockFile(dir, tasks)
	return tasks, nil
}

// loadFromLock загружает задачи из lock.json.
func loadFromLock(dir string) (map[string]Task, error) {
	data, err := os.ReadFile(filepath.Join(dir, lockFileName))
	if err != nil {
		return nil, err
	}

	var tasks map[string]Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// loadFromFiles сканирует .md файлы в директории хранилища.
func loadFromFiles(dir string) (map[string]Task, error) {
	tasks := make(map[string]Task)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".md")
		task, err := readTaskFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("parse task %s: %w", id, err)
		}
		task.ID = id
		tasks[id] = *task
	}

	return tasks, nil
}

// readTaskFile парсит .md файл в Task.
//
// Формат файла:
//
//	# Task Title
//	- Status # Draft, In Progress, Done
//	Task description (может быть многострочным)
func readTaskFile(path string) (*Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid task file: %s", path)
	}

	// Строка 1: "# Title"
	title := strings.TrimPrefix(lines[0], "# ")
	if title == lines[0] {
		return nil, fmt.Errorf("invalid title line in: %s", path)
	}

	// Строка 2: "- Status # Draft, In Progress, Done"
	statusLine := lines[1]
	statusStr := strings.TrimPrefix(statusLine, "- ")
	if statusStr == statusLine {
		return nil, fmt.Errorf("invalid status line in: %s", path)
	}
	if idx := strings.Index(statusStr, " #"); idx != -1 {
		statusStr = statusStr[:idx]
	}
	status, err := ParseStatus(strings.TrimSpace(statusStr))
	if err != nil {
		return nil, fmt.Errorf("parse status in %s: %w", path, err)
	}

	// Строки 3+: описание
	var description string
	if len(lines) > 2 {
		description = strings.Join(lines[2:], "\n")
		description = strings.TrimRight(description, "\n")
	}

	return &Task{
		Title:       title,
		Status:      status,
		Description: description,
	}, nil
}

// writeTaskFile записывает Task в .md файл.
func writeTaskFile(dir string, task Task) error {
	var sb strings.Builder
	sb.WriteString("# ")
	sb.WriteString(task.Title)
	sb.WriteString("\n- ")
	sb.WriteString(task.Status.String())
	sb.WriteString(" # Draft, In Progress, Done\n")
	if task.Description != "" {
		sb.WriteString(task.Description)
		sb.WriteString("\n")
	}

	path := filepath.Join(dir, task.ID+".md")
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// deleteTaskFile удаляет .md файл задачи.
func deleteTaskFile(dir, id string) error {
	path := filepath.Join(dir, id+".md")
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// writeLockFile записывает lock.json — быстрое представление всех задач.
func writeLockFile(dir string, tasks map[string]Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, lockFileName), data, 0644)
}
