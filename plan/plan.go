// Service plan implements a plan management REST API.
package plan

import (
	"context"
	"fmt"
)

// StorePlanRequest represents the request body for storing a plan
type StorePlanRequest struct {
	UserID   string    `json:"userId"`
	PlanJSON PlanData  `json:"planJson"`
}

// StorePlanResponse represents the response for storing a plan
type StorePlanResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// GetPlanRequest represents the request for getting a plan
type GetPlanRequest struct {
	UserID string `json:"userId"`
}

// GetPlanResponse represents the response for getting a plan
type GetPlanResponse struct {
	Plan  *Plan  `json:"plan,omitempty"`
	Error string `json:"error,omitempty"`
}

// DeletePlanRequest represents the request for deleting a plan
type DeletePlanRequest struct {
	UserID string `json:"userId"`
}

// DeletePlanResponse represents the response for deleting a plan
type DeletePlanResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

//encore:api public method=POST path=/store-plan
func StorePlan(ctx context.Context, req *StorePlanRequest) (*StorePlanResponse, error) {
	if req.UserID == "" {
		return &StorePlanResponse{
			Success: false,
			Error:   "userId is required",
		}, nil
	}

	err := InsertPlan(ctx, req.UserID, req.PlanJSON)
	if err != nil {
		return &StorePlanResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to store plan: %v", err),
		}, nil
	}

	return &StorePlanResponse{
		Success: true,
	}, nil
}

//encore:api public method=POST path=/get-plan
func GetPlan(ctx context.Context, req *GetPlanRequest) (*GetPlanResponse, error) {
	if req.UserID == "" {
		return &GetPlanResponse{
			Error: "userId is required",
		}, nil
	}

	plan, err := GetPlanByUserID(ctx, req.UserID)
	if err != nil {
		return &GetPlanResponse{
			Error: fmt.Sprintf("Failed to get plan: %v", err),
		}, nil
	}

	if plan == nil {
		return &GetPlanResponse{
			Error: "No plan found for user",
		}, nil
	}

	return &GetPlanResponse{
		Plan: plan,
	}, nil
}

//encore:api public method=POST path=/delete-plan
func DeletePlan(ctx context.Context, req *DeletePlanRequest) (*DeletePlanResponse, error) {
	if req.UserID == "" {
		return &DeletePlanResponse{
			Success: false,
			Error:   "userId is required",
		}, nil
	}

	err := DeletePlanByUserID(ctx, req.UserID)
	if err != nil {
		return &DeletePlanResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to delete plan: %v", err),
		}, nil
	}

	return &DeletePlanResponse{
		Success: true,
	}, nil
} 