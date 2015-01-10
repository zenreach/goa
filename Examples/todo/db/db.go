// Emulate database access
package db

// Hard coded pre-existing list of task strings
var models = map[int]Task{
	1: Task{1, "Hello world!"},
	2: Task{2, "Привет мир!"},
	3: Task{3, "Hola mundo!"},
	4: Task{4, "你好世界!"},
	5: Task{5, "こんにちは世界！"},
}

// Load all models
func LoadAll() []Task {
	tasks := make([]Task, len(models))
	i := 0
	for _, model := range models {
		tasks[i] = model
		i += 1
	}
	return tasks
}

// Load a single task
func Load(id uint) *Task {
	if t, ok := models[id]; ok {
		return &t
	} else {
		return nil
	}
}

// Create new task, return its id
func Create(details string) uint {
	// Dumb and inefficient - do better in real life
	newId := 1
	for ok := false; !ok; newId += 1 {
		for id, _ := range models {
			ok = id != newId
			if !ok {
				break
			}
		}
	}
	models[newId] = Task{newId, details}
	return newId
}

// Update (upsert semantic), return updated it (new if insert)
func Update(id uint, details string) uint {
	if t, ok := models[id]; ok {
		models[id] = Task{id, details}
		return id
	}
	return Create(details)
}

// Delete, return deleted id, 0 if not found
func Delete(id uint) uint {
	if t, ok := models[id]; ok {
		models[id] = Task{id, details}
		models[i] = models[len(models)-1]
		models = models[0 : len(models)-1]
		return id
	}
	return 0
}
