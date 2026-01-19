package models

import "time"

// BashScript represents a bash script stored in the database
// Script content is encrypted at rest using AES-256-GCM
type BashScript struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`             // Display name for the script
	Description string    `json:"description"`      // Optional description
	Content     string    `json:"content"`          // Script content (encrypted in DB)
	Filename    string    `json:"filename"`         // Original filename if uploaded
	Group       string    `json:"group"`            // Group/category for organization
	Source      string    `json:"source,omitempty"` // "sqlite" or "vault"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BashScriptCreate represents the data needed to create a new bash script
type BashScriptCreate struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content" validate:"required"`
	Filename    string `json:"filename,omitempty"`
	Group       string `json:"group"` // Optional, defaults to "default"
}

// BashScriptUpdate represents the data that can be updated for a bash script
type BashScriptUpdate struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content,omitempty"`
	Filename    string `json:"filename,omitempty"`
	Group       string `json:"group,omitempty"`
}

// BashScriptResponse is the API response format
type BashScriptResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content,omitempty"` // Only included when specifically requested
	Filename    string    `json:"filename"`
	Group       string    `json:"group"`            // Group/category for organization
	Source      string    `json:"source,omitempty"` // "sqlite" or "vault"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts a BashScript to a response
// includeContent controls whether the full script content is included
func (s *BashScript) ToResponse(includeContent bool) *BashScriptResponse {
	content := ""
	if includeContent {
		content = s.Content
	}
	return &BashScriptResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Content:     content,
		Filename:    s.Filename,
		Group:       s.Group,
		Source:      s.Source,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

// BashScriptsToList converts a slice of BashScripts to responses (without content)
func BashScriptsToList(scripts []*BashScript) []*BashScriptResponse {
	result := make([]*BashScriptResponse, len(scripts))
	for i, s := range scripts {
		result[i] = s.ToResponse(false)
	}
	return result
}
