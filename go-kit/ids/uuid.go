package ids

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func NewUUID() string {
	return uuid.New().String()
}

func NewUUIDWithoutDashes() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func NewUUID7() string {
	return uuid.Must(uuid.NewV7()).String()
}

func NewUUID7WithoutDashes() string {
	return strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", "")
}

func NewRefCode() string {
	refCode := uuid.New().String()
	return fmt.Sprintf("%s", refCode[len(refCode)-8:])
}
