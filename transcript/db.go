package transcript

import (
	"encore.dev/storage/sqldb"
)

// Create the transcript database and assign it to the "transcriptdb" variable
var transcriptdb = sqldb.NewDatabase("transcript", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// Course represents a single course from the transcript
type Course struct {
	Semester string `json:"semester"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Credits  string `json:"credits"`
	Grade    string `json:"grade"`
	LessonID string `json:"lesson_id,omitempty"`
}

// Transcript represents a user's transcript with courses
type Transcript struct {
	ID      int64    `json:"id"`
	UserID  string   `json:"userId"`
	Courses []Course `json:"courses"`
} 