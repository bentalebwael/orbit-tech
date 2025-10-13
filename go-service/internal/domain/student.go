package domain

import "time"

// Student represents a student entity from the backend
type Student struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Class         string    `json:"class"`
	Section       string    `json:"section"`
	Roll          int       `json:"roll"`
	Phone         string    `json:"phone"`
	Address       string    `json:"address"`
	DateOfBirth   string    `json:"dateOfBirth"`
	Gender        string    `json:"gender"`
	BloodGroup    string    `json:"bloodGroup"`
	GuardianName  string    `json:"guardianName"`
	GuardianPhone string    `json:"guardianPhone"`
	GuardianEmail string    `json:"guardianEmail"`
	AdmissionDate string    `json:"admissionDate"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
