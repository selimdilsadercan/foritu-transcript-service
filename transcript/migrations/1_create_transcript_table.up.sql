CREATE TABLE transcript (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    courses JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create an index on user_id for faster lookups
CREATE INDEX idx_transcript_user_id ON transcript(user_id);

-- Create a unique constraint to ensure one transcript per user
CREATE UNIQUE INDEX idx_transcript_user_unique ON transcript(user_id); 