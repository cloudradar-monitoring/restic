package server

import "fmt"

type HttpError struct {
	ResponseCode int
	Err          error
}

func (hr HttpError) Error() string {
	return fmt.Sprintf("Server error. Response code: %d, error: %v", hr.ResponseCode, hr.Err)
}
