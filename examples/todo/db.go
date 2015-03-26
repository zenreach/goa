package main

import "time"

// Emulate DB access
type Db map[int]*TaskModel

// Task model
type TaskModel struct {
	Id        int
	Details   string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Hard coded pre-existing list of task strings
var now = time.Now()
var expiresAt = now.Add(1 * time.Hour)
var db = Db{
	1: &TaskModel{Id: 1, Details: "Hello world!", CreatedAt: now, ExpiresAt: expiresAt},
	2: &TaskModel{Id: 2, Details: "Привет мир!", CreatedAt: now, ExpiresAt: expiresAt},
	3: &TaskModel{Id: 3, Details: "Bonjour monde!", CreatedAt: now, ExpiresAt: expiresAt},
	4: &TaskModel{Id: 4, Details: "你好世界!", CreatedAt: now, ExpiresAt: expiresAt},
	5: &TaskModel{Id: 5, Details: "こんにちは世界！", CreatedAt: now, ExpiresAt: expiresAt},
}

// Load all tasks
func (d *Db) LoadAll() []*TaskModel {
	i := 0
	res := make([]*TaskModel, len(db))
	for _, t := range db {
		res[i] = t
		i += 1
	}
	return res
}

// Load a single task
func (d *Db) Load(id int) *TaskModel {
	return db[id]
}

// Create new task, return its id
func (d *Db) Create(details string, expiresAt time.Time) int {
	// Dumb and inefficient - do better in real life
	newId := 1
	for ok := false; !ok; newId += 1 {
		for id, _ := range db {
			ok = id != newId
			if !ok {
				break
			}
		}
	}
	db[newId] = &TaskModel{Id: newId, Details: details, ExpiresAt: expiresAt}
	return newId
}

// Update (upsert semantic), return updated it (new if insert)
func (d *Db) Update(id int, details string, expiresAt time.Time) int {
	if _, ok := db[id]; ok {
		db[id] = &TaskModel{Id: id, Details: details, ExpiresAt: expiresAt}
		return id
	}
	return d.Create(details, expiresAt)
}

// Delete, return deleted id, 0 if not found
func (d *Db) Delete(id int) int {
	_, exists := db[id]
	if exists {
		delete(db, id)
		return id
	} else {
		return 0
	}
}
