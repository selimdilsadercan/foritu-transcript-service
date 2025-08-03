package plan

import (
	"context"
	"encoding/json"
	"errors"
	"encore.dev/storage/sqldb"
)

// InsertPlan inserts a new plan for a user
func InsertPlan(ctx context.Context, userID string, planJSON PlanData) error {
	planJSONBytes, err := json.Marshal(planJSON)
	if err != nil {
		return err
	}

	_, err = plandb.Exec(ctx, `
		INSERT INTO plan (user_id, plan_json)
		VALUES ($1, $2)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			plan_json = $2,
			updated_at = NOW()
	`, userID, planJSONBytes)
	
	return err
}

// GetPlanByUserID retrieves a plan for a specific user
func GetPlanByUserID(ctx context.Context, userID string) (*Plan, error) {
	var plan Plan
	var planJSONBytes []byte

	err := plandb.QueryRow(ctx, `
		SELECT id, user_id, plan_json
		FROM plan
		WHERE user_id = $1
	`, userID).Scan(&plan.ID, &plan.UserID, &planJSONBytes)

	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, nil // No plan found
		}
		return nil, err
	}

	// Parse the plan JSON
	var planJSON PlanData
	err = json.Unmarshal(planJSONBytes, &planJSON)
	if err != nil {
		return nil, err
	}

	plan.PlanJSON = planJSON
	return &plan, nil
}

// UpdatePlanByUserID updates an existing plan for a user
func UpdatePlanByUserID(ctx context.Context, userID string, planJSON PlanData) error {
	planJSONBytes, err := json.Marshal(planJSON)
	if err != nil {
		return err
	}

	result, err := plandb.Exec(ctx, `
		UPDATE plan 
		SET plan_json = $2, updated_at = NOW()
		WHERE user_id = $1
	`, userID, planJSONBytes)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return errors.New("no plan found for user")
	}

	return nil
}

// DeletePlanByUserID deletes a plan for a specific user
func DeletePlanByUserID(ctx context.Context, userID string) error {
	result, err := plandb.Exec(ctx, `
		DELETE FROM plan
		WHERE user_id = $1
	`, userID)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return errors.New("no plan found for user")
	}

	return nil
}

// GetAllPlans retrieves all plans (useful for admin purposes)
func GetAllPlans(ctx context.Context) ([]Plan, error) {
	rows, err := plandb.Query(ctx, `
		SELECT id, user_id, plan_json
		FROM plan
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []Plan
	for rows.Next() {
		var plan Plan
		var planJSONBytes []byte

		err := rows.Scan(&plan.ID, &plan.UserID, &planJSONBytes)
		if err != nil {
			return nil, err
		}

		// Parse the plan JSON
		var planJSON PlanData
		err = json.Unmarshal(planJSONBytes, &planJSON)
		if err != nil {
			return nil, err
		}

		plan.PlanJSON = planJSON
		plans = append(plans, plan)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
} 