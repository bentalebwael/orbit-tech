package handler

import (
	"fmt"
	"regexp"
)

var studentIDRegex = regexp.MustCompile(`^[0-9]{1,20}$`)

func validateStudentID(id string) error {
	if id == "" {
		return fmt.Errorf("student ID cannot be empty")
	}
	if !studentIDRegex.MatchString(id) {
		return fmt.Errorf("student ID must be numeric (1-20 digits)")
	}
	return nil
}
