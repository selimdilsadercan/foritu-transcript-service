package transcript

import (
	"context"
	"encoding/json"
	"errors"
	"encore.dev/storage/sqldb"
)

// InsertTranscript inserts a new transcript for a user
func InsertTranscript(ctx context.Context, userID string, courses []Course) error {
	coursesJSON, err := json.Marshal(courses)
	if err != nil {
		return err
	}

	_, err = transcriptdb.Exec(ctx, `
		INSERT INTO transcript (user_id, courses)
		VALUES ($1, $2)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			courses = $2,
			updated_at = NOW()
	`, userID, coursesJSON)
	
	return err
}

// GetTranscriptByUserID retrieves a transcript for a specific user
func GetTranscriptByUserID(ctx context.Context, userID string) (*Transcript, error) {
	var transcript Transcript
	var coursesJSON []byte

	err := transcriptdb.QueryRow(ctx, `
		SELECT id, user_id, courses
		FROM transcript
		WHERE user_id = $1
	`, userID).Scan(&transcript.ID, &transcript.UserID, &coursesJSON)

	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, nil // No transcript found
		}
		return nil, err
	}

	// Parse the courses JSON
	err = json.Unmarshal(coursesJSON, &transcript.Courses)
	if err != nil {
		return nil, err
	}

	return &transcript, nil
}

// UpdateTranscriptByUserID updates an existing transcript for a user
func UpdateTranscriptByUserID(ctx context.Context, userID string, courses []Course) error {
	coursesJSON, err := json.Marshal(courses)
	if err != nil {
		return err
	}

	result, err := transcriptdb.Exec(ctx, `
		UPDATE transcript 
		SET courses = $2, updated_at = NOW()
		WHERE user_id = $1
	`, userID, coursesJSON)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return errors.New("no transcript found for user")
	}

	return nil
}

// DeleteTranscriptByUserID deletes a transcript for a specific user
func DeleteTranscriptByUserID(ctx context.Context, userID string) error {
	result, err := transcriptdb.Exec(ctx, `
		DELETE FROM transcript
		WHERE user_id = $1
	`, userID)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return errors.New("no transcript found for user")
	}

	return nil
}

// GetAllTranscripts retrieves all transcripts (useful for admin purposes)
func GetAllTranscripts(ctx context.Context) ([]Transcript, error) {
	rows, err := transcriptdb.Query(ctx, `
		SELECT id, user_id, courses
		FROM transcript
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transcripts []Transcript
	for rows.Next() {
		var transcript Transcript
		var coursesJSON []byte

		err := rows.Scan(&transcript.ID, &transcript.UserID, &coursesJSON)
		if err != nil {
			return nil, err
		}

		// Parse the courses JSON
		err = json.Unmarshal(coursesJSON, &transcript.Courses)
		if err != nil {
			return nil, err
		}

		transcripts = append(transcripts, transcript)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transcripts, nil
} 