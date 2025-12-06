package models

import "time"

// EnvVariable represents an environment variable stored in the database
// Values are encrypted at rest using AES-256-GCM
type EnvVariable struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`        // Environment variable name (e.g., API_KEY)
	Value       string    `json:"value"`       // Decrypted value (encrypted in DB)
	Description string    `json:"description"` // Optional description
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EnvVariableCreate represents the data needed to create a new environment variable
type EnvVariableCreate struct {
	Name        string `json:"name" validate:"required"`
	Value       string `json:"value" validate:"required"`
	Description string `json:"description,omitempty"`
}

// EnvVariableUpdate represents the data that can be updated for an environment variable
type EnvVariableUpdate struct {
	Name        string `json:"name,omitempty"`
	Value       string `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
}

// EnvVariableResponse is the API response format (value masked by default)
type EnvVariableResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Value       string    `json:"value"` // Will be masked unless explicitly requested
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts an EnvVariable to a response with masked value
func (e *EnvVariable) ToResponse(showValue bool) *EnvVariableResponse {
	value := "••••••••"
	if showValue {
		value = e.Value
	}
	return &EnvVariableResponse{
		ID:          e.ID,
		Name:        e.Name,
		Value:       value,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// ToMap converts a slice of EnvVariables to a map for command execution
func EnvVariablesToMap(vars []*EnvVariable) map[string]string {
	result := make(map[string]string)
	for _, v := range vars {
		result[v.Name] = v.Value
	}
	return result
}
