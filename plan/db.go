package plan

import (
	"encore.dev/storage/sqldb"
)

// Create the plan database and assign it to the "plandb" variable
var plandb = sqldb.NewDatabase("plan", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// Course represents a course in the plan
type Course struct {
	Type     string   `json:"type"`
	Code     string   `json:"code,omitempty"`
	Name     string   `json:"name,omitempty"`
	Category string   `json:"category,omitempty"`
	Options  []string `json:"options,omitempty"`
}

// PlanData represents the structure of the plan JSON - array of semesters (each semester is an array of courses)
type PlanData [][]Course

// Plan represents a user's academic plan
type Plan struct {
	ID       int64     `json:"id"`
	UserID   string    `json:"userId"`
	PlanJSON PlanData  `json:"planJson"`
}

// PlanRequest represents the request body for storing a plan
type PlanRequest struct {
	UserID   string    `json:"userId"`
	PlanJSON PlanData  `json:"planJson"`
}

// PlanResponse represents the response for getting a plan
type PlanResponse struct {
	Plan  *Plan  `json:"plan,omitempty"`
	Error string `json:"error,omitempty"`
} 