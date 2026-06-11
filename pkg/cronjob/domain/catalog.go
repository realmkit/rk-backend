package domain

import "time"

const (
	// JobEventsDispatchPending dispatches pending event outbox rows.
	JobEventsDispatchPending = "events.dispatch-pending"

	// JobForumsFlushThreadViews flushes buffered forum view counters.
	JobForumsFlushThreadViews = "forums.flush-thread-views"

	// JobForumsVerifyStats verifies forum stats counters.
	JobForumsVerifyStats = "forums.verify-stats"

	// JobForumsVerifyLikes verifies forum like counters.
	JobForumsVerifyLikes = "forums.verify-likes"

	// JobPunishmentsExpireActive expires active punishments.
	JobPunishmentsExpireActive = "punishments.expire-active"

	// JobPunishmentsVerifyRestrictions verifies punishment restrictions.
	JobPunishmentsVerifyRestrictions = "punishments.verify-restrictions"

	// JobPunishmentsRebuildRestrictions rebuilds punishment restrictions.
	JobPunishmentsRebuildRestrictions = "punishments.rebuild-restrictions"

	// JobTicketsDetectSLABreaches emits events for overdue ticket work.
	JobTicketsDetectSLABreaches = "tickets.detect-sla-breaches"

	// JobTicketsCloseStale closes stale submitter-blocked tickets.
	JobTicketsCloseStale = "tickets.close-stale"

	// JobTicketsVerifyStats verifies ticket counters.
	JobTicketsVerifyStats = "tickets.verify-stats"

	// JobTicketsRebuildStats rebuilds ticket counters.
	JobTicketsRebuildStats = "tickets.rebuild-stats"

	// JobAssetsExpireUploadIntents expires stale upload intents.
	JobAssetsExpireUploadIntents = "assets.expire-upload-intents"

	// JobUsersCleanupIdentityClaims cleans stale identity claims.
	JobUsersCleanupIdentityClaims = "users.cleanup-identity-claims"
)

// DefaultDefinitions returns the initial code-owned job definitions.
func DefaultDefinitions(now time.Time) []Definition {
	return []Definition{
		interval(JobEventsDispatchPending, "Dispatch pending events", time.Minute, now),
		interval(JobForumsFlushThreadViews, "Flush forum thread views", 5*time.Minute, now),
		interval(JobForumsVerifyStats, "Verify forum stats", 24*time.Hour, now),
		interval(JobForumsVerifyLikes, "Verify forum likes", 24*time.Hour, now),
		interval(JobPunishmentsExpireActive, "Expire active punishments", 5*time.Minute, now),
		interval(JobPunishmentsVerifyRestrictions, "Verify punishment restrictions", 24*time.Hour, now),
		manual(JobPunishmentsRebuildRestrictions, "Rebuild punishment restrictions", now),
		interval(JobTicketsDetectSLABreaches, "Detect ticket SLA breaches", 15*time.Minute, now),
		interval(JobTicketsCloseStale, "Close stale tickets", 24*time.Hour, now),
		interval(JobTicketsVerifyStats, "Verify ticket stats", 24*time.Hour, now),
		manual(JobTicketsRebuildStats, "Rebuild ticket stats", now),
		interval(JobAssetsExpireUploadIntents, "Expire upload intents", 30*time.Minute, now),
		interval(JobUsersCleanupIdentityClaims, "Cleanup identity claims", 24*time.Hour, now),
	}
}

// interval creates an interval definition.
func interval(key string, name string, duration time.Duration, now time.Time) Definition {
	next := now.Add(duration)
	return Definition{
		Key:                key,
		Name:               name,
		ScheduleKind:       ScheduleInterval,
		ScheduleExpression: duration.String(),
		Enabled:            true,
		ConcurrencyPolicy:  ConcurrencyForbid,
		NextRunAt:          &next,
		Version:            1,
	}
}

// manual creates a disabled manual definition.
func manual(key string, name string, now time.Time) Definition {
	return Definition{
		Key:                key,
		Name:               name,
		ScheduleKind:       ScheduleManual,
		ScheduleExpression: "",
		Enabled:            false,
		ConcurrencyPolicy:  ConcurrencyForbid,
		Version:            1,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}
