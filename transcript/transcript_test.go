package transcript

import (
	"context"
	"testing"
)

// TestParseTranscript tests the API endpoint
func TestParseTranscript(t *testing.T) {
	ctx := context.Background()
	req := &ParseTranscriptRequest{}
	
	resp, err := ParseTranscript(ctx, req)
	if err != nil {
		t.Fatalf("Failed to call ParseTranscript: %v", err)
	}
	
	// Verify response structure
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}
	
	if len(resp.Courses) == 0 {
		t.Error("Expected at least one course in response")
	}
	
	// Verify first course details
	if len(resp.Courses) > 0 {
		course := resp.Courses[0]
		if course.Semester != "2021-2022 Bahar Dönemi" {
			t.Errorf("Expected semester '2021-2022 Bahar Dönemi', got '%s'", course.Semester)
		}
		if course.Code != "BLG 102E" {
			t.Errorf("Expected code 'BLG 102E', got '%s'", course.Code)
		}
		if course.Name != "Intr to Sci&Eng Comp)" {
			t.Errorf("Expected name 'Intr to Sci&Eng Comp)', got '%s'", course.Name)
		}
		if course.Credits != "4" {
			t.Errorf("Expected credits '4', got '%s'", course.Credits)
		}
		if course.Grade != "CC" {
			t.Errorf("Expected grade 'CC', got '%s'", course.Grade)
		}
	}
	
	// Verify error message indicates placeholder
	if resp.Error == "" {
		t.Error("Expected error message indicating placeholder implementation")
	}
} 