package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Response struct {
	RequestID    string      `json:"request_id,omitempty"`
	Success      bool        `json:"success"`
	ResponseCode int         `json:"response_code"`
	Data         interface{} `json:"data,omitempty"`
	Message      string      `json:"message"`
}

func PrepareSuccessResponse(message string, responseCode int, data interface{}) Response {
	return Response{
		Message:      message,
		ResponseCode: responseCode,
		Data:         data,
		Success:      true,
	}
}
func PrepareErrorResponse(message string, responseCode int) Response {
	return Response{
		Message:      message,
		ResponseCode: responseCode,
		Success:      false,
	}
}

func ParseDuration(durationStr string) (time.Duration, error) {
	parts := strings.Split(durationStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid duration format")
	}

	numberStr := parts[0]
	unitStr := parts[1]

	number, err := strconv.ParseFloat(numberStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration number")
	}

	var duration time.Duration
	switch strings.ToLower(unitStr) {
	case "s":
		duration = time.Duration(number * float64(time.Second))
	case "m":
		duration = time.Duration(number * float64(time.Minute))
	case "h":
		duration = time.Duration(number * float64(time.Hour))
	default:
		return 0, fmt.Errorf("invalid duration unit")
	}

	return duration, nil
}

func GenerateRandomHex(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}
func ComparePassword(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetCredentials(credentials string) (string, string, error) {
	splitCredentials := strings.Split(credentials, ":")
	if len(splitCredentials) != 2 {
		return "", "", fmt.Errorf("invalid credentials passed")
	}

	moduleName := splitCredentials[0]
	password := splitCredentials[1]

	return moduleName, password, nil
}
