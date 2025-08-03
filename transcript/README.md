# Transcript Service

This service provides functionality to parse PDF transcripts and store them in a PostgreSQL database using Encore's SQL database primitives.

## Database Schema

The service uses a `transcript` table with the following structure:

```sql
CREATE TABLE transcript (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    courses JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## API Endpoints

### 1. Parse and Store Transcript (Combined)
**POST** `/parse-and-store-transcript`

Combines parsing and storing functionality in a single endpoint. Accepts a PDF file (base64 encoded) and a user ID, then parses the transcript and stores it in the database.

**Request:**
```json
{
    "userId": "user123",
    "pdf_base64": "base64_encoded_pdf_content"
}
```

**Response:**
```json
{
    "transcript": {
        "id": 1,
        "userId": "user123",
        "courses": [
            {
                "semester": "2021-2022 Bahar Dönemi",
                "code": "BLG 102E",
                "name": "Intr to Sci&Eng Comp)",
                "credits": "4",
                "grade": "CC"
            }
        ]
    },
    "debug": "Debug information from parsing process"
}
```

### 2. Store Transcript
**POST** `/transcript`

Stores a transcript with pre-parsed course data.

**Request:**
```json
{
    "userId": "user123",
    "courses": [
        {
            "semester": "2021-2022 Bahar Dönemi",
            "code": "BLG 102E",
            "name": "Intr to Sci&Eng Comp)",
            "credits": "4",
            "grade": "CC"
        }
    ]
}
```

**Response:**
```json
{
    "message": "Transcript stored successfully",
    "userId": "user123"
}
```

### 3. Get Transcript
**GET** `/transcript/:userID`

Retrieves a transcript for a specific user.

**Response:**
```json
{
    "transcript": {
        "id": 1,
        "userId": "user123",
        "courses": [...]
    }
}
```

### 4. Update Transcript
**PUT** `/transcript/:userID`

Updates an existing transcript.

**Request:**
```json
{
    "courses": [...]
}
```

**Response:**
```json
{
    "message": "Transcript updated successfully",
    "userId": "user123"
}
```

### 5. Delete Transcript
**DELETE** `/transcript/:userID`

Deletes a transcript for a specific user.

**Response:**
```json
{
    "message": "Transcript deleted successfully",
    "userId": "user123"
}
```

### 6. List All Transcripts
**GET** `/transcripts`

Retrieves all transcripts in the database.

**Response:**
```json
{
    "transcripts": [...],
    "count": 5
}
```

### 7. Parse Transcript Only
**POST** `/parse-transcript`

Parses a PDF transcript without storing it in the database.

**Request:**
```json
{
    "pdf_base64": "base64_encoded_pdf_content"
}
```

**Response:**
```json
{
    "courses": [...],
    "debug": "Debug information from parsing process"
}
```

## Database Operations

### Insert Transcript
```go
err := transcript.InsertTranscript(ctx, "user123", courses)
```

### Get Transcript
```go
transcript, err := transcript.GetTranscriptByUserID(ctx, "user123")
```

### Update Transcript
```go
err := transcript.UpdateTranscriptByUserID(ctx, "user123", courses)
```

### Delete Transcript
```go
err := transcript.DeleteTranscriptByUserID(ctx, "user123")
```

### Get All Transcripts
```go
transcripts, err := transcript.GetAllTranscripts(ctx)
```

## Utility Functions

### Load Courses from JSON File
```go
courses, err := transcript.LoadCoursesFromJSONFile("transcript_simple.json")
```

### Convert Raw JSON to Courses
```go
courses, err := transcript.ConvertRawJSONToCourses(rawJSON)
```

### Filter Courses by Semester
```go
filteredCourses := transcript.FilterCoursesBySemester(courses, "2021-2022 Bahar Dönemi")
```

### Filter Courses by Grade
```go
filteredCourses := transcript.FilterCoursesByGrade(courses, "AA")
```

### Calculate GPA
```go
gpa := transcript.CalculateGPA(courses)
```

## Usage Examples

### Combined Parse and Store
```go
package main

import (
    "context"
    "foritu/transcript"
)

func main() {
    ctx := context.Background()
    
    // Load PDF and encode to base64
    pdfBase64 := "your_base64_encoded_pdf"
    
    req := &transcript.ParseAndStoreTranscriptRequest{
        UserID:    "user123",
        PDFBase64: pdfBase64,
    }
    
    resp, err := transcript.ParseAndStoreTranscript(ctx, req)
    if err != nil {
        // Handle error
    }
    
    if resp.Transcript != nil {
        fmt.Printf("Stored transcript with %d courses\n", len(resp.Transcript.Courses))
    }
}
```

### Store Pre-parsed Data
```go
package main

import (
    "context"
    "foritu/transcript"
)

func main() {
    ctx := context.Background()
    
    // Load courses from JSON file
    courses, err := transcript.LoadCoursesFromJSONFile("transcript_simple.json")
    if err != nil {
        log.Fatal(err)
    }
    
    // Store in database
    err = transcript.InsertTranscript(ctx, "user123", courses)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Transcript stored successfully")
}
```

## Running the Application

1. Make sure you have Encore CLI installed
2. Run the application:
   ```bash
   encore run
   ```

3. The application will automatically:
   - Pull the PostgreSQL Docker image
   - Create the database and tables
   - Start the API server

## Testing

You can test the API endpoints using the provided test scripts:

- `test_transcript_db.go` - Tests basic database operations
- `test_parse_and_store.go` - Tests the combined parse and store functionality

## Error Handling

The service uses Encore's error handling primitives and returns appropriate HTTP status codes:

- `400 Bad Request` - Invalid input parameters
- `404 Not Found` - Transcript not found
- `500 Internal Server Error` - Database or parsing errors

## Data Types

### Course
```go
type Course struct {
    Semester string `json:"semester"`
    Code     string `json:"code"`
    Name     string `json:"name"`
    Credits  string `json:"credits"`
    Grade    string `json:"grade"`
}
```

### Transcript
```go
type Transcript struct {
    ID      int64    `json:"id"`
    UserID  string   `json:"userId"`
    Courses []Course `json:"courses"`
}
```

### TranscriptCourse (for parsing)
```go
type TranscriptCourse struct {
    Semester string `json:"semester"`
    Code     string `json:"code"`
    Name     string `json:"name"`
    Credits  string `json:"credits"`
    Grade    string `json:"grade"`
} 