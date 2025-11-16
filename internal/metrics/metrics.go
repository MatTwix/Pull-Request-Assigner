package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// PR metrics
	PRCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pr_created_total",
			Help: "Total number of created PR's",
		},
	)
	PRMerged = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pr_merged_total",
			Help: "Total number of merged PR's",
		},
	)

	// User metrics
	UsersCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_created_total",
			Help: "Total number of created users",
		},
	)
	UserStatusChanges = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_status_changes_total",
			Help: "Any changes of is_active field",
		},
		[]string{"operation"},
	)

	// Team metrics
	TeamsCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "teams_created_total",
			Help: "Total numbеr of created teams",
		},
	)

	// Other metrics
	BusinessErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "business_errors_total",
			Help: "Logic errors",
		},
		[]string{"type"},
	)
)
