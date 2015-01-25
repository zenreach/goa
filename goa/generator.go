package main

import (
	"fmt"
	"go/ast"
)

// Generate JSON schema from arbitrary data structure.
// Struct field tags may be used to specify validation rules.
func generateJsonSchema(st *ast.StructType) (map[string]interface{}, error) {
	fields := st.Fields.List
	for _, field := range fields {
		typ := field.Type
		fmt.Printf("Type: %v+\n", typ)
		for _, name := range field.Names {
			fmt.Printf("Name: %v+\n", name)
		}
	}
	return map[string]interface{}{}, nil
}

// Json schema defining single data type
func typeSchema(t string) map[string]interface{} {
	return map[string]interface{}{"type": t}
}
