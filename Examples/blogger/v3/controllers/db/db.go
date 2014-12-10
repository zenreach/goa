// Emulate database access, keep all data in memory
package db

import "../../../models"

// Retrieve author by id
func getAuthor(id) model.Author {
	return authors[id]
}

// Initial data

var authors = map[string]model.Author{
	"1": model.Author{},
}
