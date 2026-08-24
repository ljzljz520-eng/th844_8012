package audit

import (
	"fmt"
	"time"

	"agri-packaging/internal/model"
)

func HumanLabel(event model.AuditEvent) string {
	when := event.CreatedAt.Format(time.RFC3339)
	if event.Detail == "" {
		return fmt.Sprintf("%s %s %s", when, event.Actor, event.Action)
	}
	return fmt.Sprintf("%s %s %s (%s)", when, event.Actor, event.Action, event.Detail)
}
