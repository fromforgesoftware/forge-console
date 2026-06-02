package app

import (
	"context"
	"log/slog"
)

// logActionAuditor records cluster actions to the structured log. It satisfies
// ActionAuditor as the default sink; a hallmark-backed auditor can replace it
// without changing the usecase.
type logActionAuditor struct {
	logger *slog.Logger
}

// NewLogActionAuditor returns the default logging auditor.
func NewLogActionAuditor() ActionAuditor {
	return &logActionAuditor{logger: slog.Default()}
}

func (a *logActionAuditor) RecordTopologyAction(_ context.Context, actor, verb, target string) {
	a.logger.Info("platform.action", "actor", actor, "verb", verb, "target", target)
}
