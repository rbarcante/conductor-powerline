package segments

// Usage threshold constants shared across block and weekly segments.
// Color changes from normal → warning → critical as usage increases.
const (
	usageWarningThreshold  = 70.0 // percent — switches to warning colors
	usageCriticalThreshold = 90.0 // percent — switches to critical colors
)

// Context window threshold constants for the context segment.
const (
	contextWarningThreshold  = 50 // percent of context window used
	contextCriticalThreshold = 80 // percent of context window used
)
