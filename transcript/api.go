package transcript

import (
	"context"
	"encore.dev/beta/errs"
	"fmt"
)

//encore:api public method=POST path=/transcript
func StoreTranscript(ctx context.Context, req *StoreTranscriptRequest) (*StoreTranscriptResponse, error) {
	if req.UserID == "" {
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: "user_id is required",
		}
	}

	if len(req.Courses) == 0 {
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: "courses cannot be empty",
		}
	}

	err := InsertTranscript(ctx, req.UserID, req.Courses)
	if err != nil {
		return nil, &errs.Error{
			Code: errs.Internal,
			Message: "failed to store transcript",
		}
	}

	return &StoreTranscriptResponse{
		Message: "Transcript stored successfully",
		UserID:  req.UserID,
	}, nil
}

//encore:api public method=GET path=/transcript/:userID
func GetTranscript(ctx context.Context, userID string) (*GetTranscriptResponse, error) {
	if userID == "" {
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: "user_id is required",
		}
	}

	transcript, err := GetTranscriptByUserID(ctx, userID)
	if err != nil {
		return nil, &errs.Error{
			Code: errs.Internal,
			Message: "failed to retrieve transcript",
		}
	}

	if transcript == nil {
		return nil, &errs.Error{
			Code: errs.NotFound,
			Message: "transcript not found",
		}
	}

	return &GetTranscriptResponse{
		Transcript: transcript,
	}, nil
}

//encore:api public method=PUT path=/transcript/:userID
func UpdateTranscript(ctx context.Context, userID string, req *UpdateTranscriptRequest) (*UpdateTranscriptResponse, error) {
	if userID == "" {
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: "user_id is required",
		}
	}

	if len(req.Courses) == 0 {
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: "courses cannot be empty",
		}
	}

	err := UpdateTranscriptByUserID(ctx, userID, req.Courses)
	if err != nil {
		return nil, &errs.Error{
			Code: errs.Internal,
			Message: "failed to update transcript",
		}
	}

	return &UpdateTranscriptResponse{
		Message: "Transcript updated successfully",
		UserID:  userID,
	}, nil
}

//encore:api public method=DELETE path=/transcript/:userID
func DeleteTranscript(ctx context.Context, userID string) (*DeleteTranscriptResponse, error) {
	if userID == "" {
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
			Message: "user_id is required",
		}
	}

	err := DeleteTranscriptByUserID(ctx, userID)
	if err != nil {
		return nil, &errs.Error{
			Code: errs.Internal,
			Message: "failed to delete transcript",
		}
	}

	return &DeleteTranscriptResponse{
		Message: "Transcript deleted successfully",
		UserID:  userID,
	}, nil
}

//encore:api public method=GET path=/transcripts
func ListAllTranscripts(ctx context.Context) (*ListTranscriptsResponse, error) {
	transcripts, err := GetAllTranscripts(ctx)
	if err != nil {
		return nil, &errs.Error{
			Code: errs.Internal,
			Message: "failed to retrieve transcripts",
		}
	}

	return &ListTranscriptsResponse{
		Transcripts: transcripts,
		Count:       len(transcripts),
	}, nil
}

// ParseAndStoreTranscriptRequest represents the request for parsing and storing a transcript
type ParseAndStoreTranscriptRequest struct {
	UserID   string `json:"userId"`
	PDFBase64 string `json:"pdf_base64"`
}

// ParseAndStoreTranscriptResponse represents the response
type ParseAndStoreTranscriptResponse struct {
	Transcript *Transcript `json:"transcript,omitempty"`
	Error      string      `json:"error,omitempty"`
	Debug      string      `json:"debug,omitempty"`
}

//encore:api public method=POST path=/parse-and-store-transcript
func ParseAndStoreTranscript(ctx context.Context, req *ParseAndStoreTranscriptRequest) (*ParseAndStoreTranscriptResponse, error) {
	if req.UserID == "" {
		return &ParseAndStoreTranscriptResponse{
			Error: "userId is required",
		}, nil
	}

	if req.PDFBase64 == "" {
		return &ParseAndStoreTranscriptResponse{
			Error: "pdf_base64 is required",
		}, nil
	}

	// First, parse the transcript using the existing parsing logic
	parseReq := &ParseTranscriptRequest{
		PDFBase64: req.PDFBase64,
	}

	parseResp, err := ParseTranscript(ctx, parseReq)
	if err != nil {
		return &ParseAndStoreTranscriptResponse{
			Error: fmt.Sprintf("Failed to parse transcript: %v", err),
		}, nil
	}

	if parseResp.Error != "" {
		return &ParseAndStoreTranscriptResponse{
			Error: parseResp.Error,
			Debug: parseResp.Debug,
		}, nil
	}

	// Convert TranscriptCourse to Course for database storage
	var courses []Course
	for _, tc := range parseResp.Courses {
		courses = append(courses, Course{
			Semester: tc.Semester,
			Code:     tc.Code,
			Name:     tc.Name,
			Credits:  tc.Credits,
			Grade:    tc.Grade,
		})
	}

	// Store the parsed transcript in the database
	err = InsertTranscript(ctx, req.UserID, courses)
	if err != nil {
		return &ParseAndStoreTranscriptResponse{
			Error: fmt.Sprintf("Failed to store transcript: %v", err),
			Debug: parseResp.Debug,
		}, nil
	}

	// Retrieve the stored transcript to return
	storedTranscript, err := GetTranscriptByUserID(ctx, req.UserID)
	if err != nil {
		return &ParseAndStoreTranscriptResponse{
			Error: fmt.Sprintf("Failed to retrieve stored transcript: %v", err),
			Debug: parseResp.Debug,
		}, nil
	}

	return &ParseAndStoreTranscriptResponse{
		Transcript: storedTranscript,
		Debug:      parseResp.Debug,
	}, nil
}

// Request and Response types
type StoreTranscriptRequest struct {
	UserID  string   `json:"userId"`
	Courses []Course `json:"courses"`
}

type StoreTranscriptResponse struct {
	Message string `json:"message"`
	UserID  string `json:"userId"`
}

type GetTranscriptResponse struct {
	Transcript *Transcript `json:"transcript"`
}

type UpdateTranscriptRequest struct {
	Courses []Course `json:"courses"`
}

type UpdateTranscriptResponse struct {
	Message string `json:"message"`
	UserID  string `json:"userId"`
}

type DeleteTranscriptResponse struct {
	Message string `json:"message"`
	UserID  string `json:"userId"`
}

type ListTranscriptsResponse struct {
	Transcripts []Transcript `json:"transcripts"`
	Count       int          `json:"count"`
} 