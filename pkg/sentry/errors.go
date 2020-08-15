package sentry

import (
	"fmt"
)

type APIError map[string]interface{}

func (e APIError) Error() string {
	if detail, ok := e["detail"].(string); ok {
		return fmt.Sprintf("sentry: %s", detail)
	}

	return fmt.Sprintf("sentry: %v", map[string]interface{}(e))
}
