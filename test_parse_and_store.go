package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"

	"foritu/transcript"
)

func main() {
	// Load a sample PDF file (you would need to provide a real PDF file)
	// For demonstration, we'll use a base64 encoded PDF or create a mock one
	pdfBase64 := getSamplePDFBase64()

	// Create a context
	ctx := context.Background()

	// Example user ID
	userID := "user456"

	fmt.Printf("Testing parse and store functionality for user: %s\n", userID)

	// Call the new combined parse and store endpoint
	req := &transcript.ParseAndStoreTranscriptRequest{
		UserID:    userID,
		PDFBase64: pdfBase64,
	}

	resp, err := transcript.ParseAndStoreTranscript(ctx, req)
	if err != nil {
		log.Fatalf("Failed to parse and store transcript: %v", err)
	}

	if resp.Error != "" {
		fmt.Printf("Error: %s\n", resp.Error)
		if resp.Debug != "" {
			fmt.Printf("Debug info: %s\n", resp.Debug)
		}
		return
	}

	if resp.Transcript != nil {
		fmt.Printf("Successfully parsed and stored transcript for user: %s\n", userID)
		fmt.Printf("Transcript ID: %d\n", resp.Transcript.ID)
		fmt.Printf("Number of courses: %d\n", len(resp.Transcript.Courses))
		
		// Display first few courses
		for i, course := range resp.Transcript.Courses {
			if i >= 5 { // Show only first 5 courses
				break
			}
			fmt.Printf("  Course %d: %s - %s (%s credits, Grade: %s)\n", 
				i+1, course.Code, course.Name, course.Credits, course.Grade)
		}
	}

	// Verify the transcript was stored by retrieving it
	fmt.Printf("\nVerifying stored transcript...\n")
	storedTranscript, err := transcript.GetTranscriptByUserID(ctx, userID)
	if err != nil {
		log.Fatalf("Failed to retrieve stored transcript: %v", err)
	}

	if storedTranscript != nil {
		fmt.Printf("✓ Transcript successfully stored and retrieved\n")
		fmt.Printf("  User ID: %s\n", storedTranscript.UserID)
		fmt.Printf("  Total courses: %d\n", len(storedTranscript.Courses))
	} else {
		fmt.Printf("✗ Transcript not found after storage\n")
	}
}

// getSamplePDFBase64 returns a sample base64 encoded PDF for testing
// In a real scenario, you would load an actual PDF file
func getSamplePDFBase64() string {
	// This is a placeholder - you would need to provide a real PDF file
	// For testing purposes, you can:
	// 1. Load a real PDF file and encode it to base64
	// 2. Use the existing transcript_simple.json to test the database operations
	
	// Option 1: Load a real PDF file (uncomment and modify path as needed)
	/*
	pdfBytes, err := ioutil.ReadFile("path/to/your/transcript.pdf")
	if err != nil {
		log.Fatalf("Failed to read PDF file: %v", err)
	}
	return base64.StdEncoding.EncodeToString(pdfBytes)
	*/
	
	// Option 2: For now, return an empty string to test the error handling
	// You can replace this with actual base64 encoded PDF content
	return ""
} 