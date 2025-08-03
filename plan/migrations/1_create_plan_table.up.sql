CREATE TABLE plan (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    plan_json JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create an index on user_id for faster lookups
CREATE INDEX idx_plan_user_id ON plan(user_id);

-- Create a unique constraint to ensure one plan per user
CREATE UNIQUE INDEX idx_plan_user_unique ON plan(user_id); 