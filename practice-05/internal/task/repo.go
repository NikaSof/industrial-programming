package task

import (
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Repo struct {
	mu    sync.RWMutex
	seq   int64
	items map[int64]Task
}

func NewRepo() *Repo {
	return &Repo{
		items: make(map[int64]Task),
	}
}

func (r *Repo) List() []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.items))
	for _, task := range r.items {
		tasks = append(tasks, task)
	}

	return tasks
}

func (r *Repo) Get(id int64) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.items[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	return task, nil
}

func (r *Repo) Create(title string) Task {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	now := time.Now().UTC()

	task := Task{
		ID:        r.seq,
		Title:     title,
		Done:      false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	r.items[task.ID] = task
	return task
}

func (r *Repo) Update(id int64, title string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.items[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	task.Title = title
	task.Done = done
	task.UpdatedAt = time.Now().UTC()

	// Из map получена копия: сохраняем изменённое значение обратно.
	r.items[id] = task

	return task, nil
}

func (r *Repo) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return ErrNotFound
	}

	delete(r.items, id)
	return nil
}
