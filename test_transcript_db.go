package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"foritu/transcript"
)

func main() {
	// Load courses from the existing JSON file
	courses, err := transcript.LoadCoursesFromJSONFile("transcript_simple.json")
	if err != nil {
		log.Fatalf("Failed to load courses: %v", err)
	}

	fmt.Printf("Loaded %d courses from JSON file\n", len(courses))

	// Create a context
	ctx := context.Background()

	// Example user ID
	userID := "user123"

	// Store the transcript in the database
	err = transcript.InsertTranscript(ctx, userID, courses)
	if err != nil {
		log.Fatalf("Failed to store transcript: %v", err)
	}

	fmt.Printf("Successfully stored transcript for user: %s\n", userID)

	// Retrieve the transcript from the database
	storedTranscript, err := transcript.GetTranscriptByUserID(ctx, userID)
	if err != nil {
		log.Fatalf("Failed to retrieve transcript: %v", err)
	}

	if storedTranscript != nil {
		fmt.Printf("Retrieved transcript for user: %s\n", storedTranscript.UserID)
		fmt.Printf("Number of courses: %d\n", len(storedTranscript.Courses))

		// Calculate GPA summary
		gpa, totalCredits, courseCount := transcript.CalculateGPASummary(storedTranscript.Courses)
		fmt.Printf("GPA: %.2f\n", gpa)
		fmt.Printf("Total Credits: %.1f\n", totalCredits)
		fmt.Printf("Course Count: %d\n", courseCount)

		// Example: Get courses by semester
		semesterCourses := transcript.GetCoursesBySemester(storedTranscript.Courses, "2021-2022 Bahar Dönemi")
		fmt.Printf("Courses in 2021-2022 Bahar Dönemi: %d\n", len(semesterCourses))

		// Example: Get courses by grade
		gradeACourses := transcript.GetCoursesByGrade(storedTranscript.Courses, "AA")
		fmt.Printf("Courses with AA grade: %d\n", len(gradeACourses))
	}

	// Example: List all transcripts (for admin purposes)
	allTranscripts, err := transcript.GetAllTranscripts(ctx)
	if err != nil {
		log.Fatalf("Failed to get all transcripts: %v", err)
	}

	fmt.Printf("Total transcripts in database: %d\n", len(allTranscripts))
} 