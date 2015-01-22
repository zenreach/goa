package main

// Emulate DB access
type Db map[uint]TaskModel

// Task model
type TaskModel struct {
	id      uint
	details string
}

// Implements TaskData
func (t *TaskModel) Id() uint        { return t.id }
func (t *TaskModel) Details() string { return t.details }

// Hard coded pre-existing list of task strings
var db = Db{
	1: TaskModel{id: 1, details: "Hello world!"},
	2: TaskModel{id: 2, details: "Привет мир!"},
	3: TaskModel{id: 3, details: "Hola mundo!"},
	4: TaskModel{id: 4, details: "你好世界!"},
	5: TaskModel{id: 5, details: "こんにちは世界！"},
}

// Load all tasks
func (d *Db) LoadAll() []TaskModel {
	tasks := make([]TaskModel, len(db))
	i := 0
	for _, model := range db {
		tasks[i] = model
		i += 1
	}
	return tasks
}

// Load a single task
func (d *Db) Load(id uint) *TaskModel {
	if t, ok := db[id]; ok {
		return &t
	} else {
		return nil
	}
}

// Create new task, return its id
func (d *Db) Create(details string) uint {
	// Dumb and inefficient - do better in real life
	newId := uint(1)
	for ok := false; !ok; newId += 1 {
		for id, _ := range db {
			ok = id != newId
			if !ok {
				break
			}
		}
	}
	db[newId] = TaskModel{id: newId, details: details}
	return newId
}

// Update (upsert semantic), return updated it (new if insert)
func (d *Db) Update(id uint, details string) uint {
	if _, ok := db[id]; ok {
		db[id] = TaskModel{id: id, details: details}
		return id
	}
	return d.Create(details)
}

// Delete, return deleted id, 0 if not found
func (d *Db) Delete(id uint) uint {
	_, exists := db[id]
	delete(db, id)
	if exists {
		return id
	} else {
		return 0
	}
}
