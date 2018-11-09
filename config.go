package addressbook

import "time"

var (
	// Version is an app version
	Version string

	// Revision is an app revision
	Revision string

	// Env is current environment
	Env string

	// Status reports application status
	Status string

	// StatusCode reports application status code
	StatusCode Code

	// StartedTime
	StartedTime time.Time

	// RequestsCount
	RequestsCount uint64
)

// Code is a type for enums
type Code uint8

// Enums for code
const (
	// In case this var was not set up, status will be unknown.
	Unknown Code = iota
	// In case of some problems (TODO: make more detailed in futute).
	HaveProblems
	// Everything is okay.
	Running
)

// CodeToText map stores text description of code enums.
var CodeToText = map[Code]string{
	Unknown:      "UNKNOWN",
	HaveProblems: "HAVE PROBLEMS",
	Running:      "RUNNIN",
}

// GetCodeText returns text description of code.
func GetCodeText(c Code) string {
	return CodeToText[c]
}

//StatusReport for server reporting page
type StatusReport struct {
	Version       string `json:"version"`                  // Version is an app version
	Revision      string `json:"revision"`                 // Revision is an app revision
	Env           string `json:"env"`                      // Env is current environment
	Status        string `json:"status"`                   // Status reports application status
	StatusCode    Code   `json:"status_code"`              // StatusCode reports application status code
	StartedTime   string `json:"started_time,omitempty"`   // StartedTime
	RequestsCount uint64 `json:"requests_count,omitempty"` // RequestsCount
}

// MakeReport creates new struct filled with actual values.
func MakeReport() StatusReport {
	return StatusReport{
		Version:       Version,
		Revision:      Revision,
		Env:           Env,
		Status:        Status,
		StatusCode:    StatusCode,
		StartedTime:   StartedTime.Format(time.RFC3339),
		RequestsCount: RequestsCount,
	}
}
