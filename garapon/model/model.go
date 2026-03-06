// Package model defines the pure data types shared across all layers.
package model

import "time"

// PrizeGrade represents the rank of a prize.
type PrizeGrade string

const (
	GradeTokutou PrizeGrade = "特等"
	GradeIttou   PrizeGrade = "1等"
	GradeNittou  PrizeGrade = "2等"
	GradeSantou  PrizeGrade = "3等"
	GradeYontou  PrizeGrade = "4等"
	GradeHazure  PrizeGrade = "参加賞"
)

// BallColor represents the visual color of a lottery ball.
type BallColor struct {
	Name string `json:"name"`
	Hex  string `json:"hex"`
}

// Prize represents a prize entry including its lottery weight.
// Weight is an integer out of 1000 (e.g. Weight=5 means 0.5% probability).
type Prize struct {
	Grade       PrizeGrade `json:"grade"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Ball        BallColor  `json:"ball"`
	Weight      int        `json:"weight"`
}

// DrawResult is returned by a single lottery draw.
type DrawResult struct {
	Prize     Prize     `json:"prize"`
	DrawnAt   time.Time `json:"drawn_at"`
	TicketNum int       `json:"ticket_num"`
}

// Stats holds aggregate information about all draws so far.
type Stats struct {
	TotalDraws  int            `json:"total_draws"`
	GradeCount  map[string]int `json:"grade_count"`
	LastUpdated time.Time      `json:"last_updated"`
}

// PrizesInfo wraps the current prize table with rotation metadata.
type PrizesInfo struct {
	Prizes              []Prize   `json:"prizes"`
	NextRotationAt      time.Time `json:"next_rotation_at"`
	LastRotatedAt       time.Time `json:"last_rotated_at"`
	RotationIntervalSec int       `json:"rotation_interval_sec"`
}

// ErrorResponse is the JSON body returned on API errors.
type ErrorResponse struct {
	Error string `json:"error"`
}
