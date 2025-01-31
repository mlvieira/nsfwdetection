package models

// ErrorResponse is a common structure for API errors
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

// PaginatedResponse defines the JSON structure for paginated API responses
type PaginatedResponse struct {
	Data  []UploadedImage `json:"data"`
	Count int             `json:"count"`
	Total int             `json:"total"`
}

type PaginatedRequest struct {
	ID       int   `json:"id"`
	Limit    int   `json:"limit"`
	Reviewed *bool `json:"reviewed,omitempty"`
}

type LabelRequest struct {
	Event  string `json:"event"`
	Sha256 string `json:"sha256"`
	Rating string `json:"rating"`
}

type AckResponse struct {
	Event  string `json:"event"`
	Sha256 string `json:"sha256"`
	Status string `json:"status"`
}

type StatsResponse struct {
	TotalImages        int            `json:"total_images"`
	ReviewedImages     int            `json:"reviewed_images"`
	UnlabeledImages    int            `json:"unlabeled_images"`
	AverageConfidence  float64        `json:"average_confidence"`
	LabelDistribution  map[string]int `json:"label_distribution"`
	LabelingEfficiency float64        `json:"labeling_efficiency_percentage"`
}
