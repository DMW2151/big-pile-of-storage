package cortx

import "time"

const (
	bucketCreatedTimeout                          = 2 * time.Minute
	bucketVersioningStableTimeout                 = 1 * time.Minute
	propagationTimeout                            = 1 * time.Minute
	lifecycleConfigurationExtraRetryDelay         = 5 * time.Second
	lifecycleConfigurationRulesPropagationTimeout = 3 * time.Minute
	lifecycleConfigurationRulesSteadyTimeout      = 2 * time.Minute
	lifecycleConfigurationRulesStatusReady        = "READY"
	lifecycleConfigurationRulesStatusNotReady     = "NOT_READY"
)
