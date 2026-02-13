package domain

type MonitoringStatus uint

const (
	MonitoringStatusUnknown MonitoringStatus = iota
	// Monitor all episodes (past and future).
	MonitoringStatusAll
	// Only monitor episodes that air after monitoring started.
	MonitoringStatusFuture
	// Monitor episodes that don't have a local file.
	MonitoringStatusMissing
	// Only monitor episodes that already have a file (for upgrades).
	MonitoringStatusExisting
	// Only monitor the first season.
	MonitoringStatusFirstSeason
	// Only monitor the most recent season.
	MonitoringStatusLatestSeason
	// Add the series, but monitor nothing.
	MonitoringStatusNone
	monitoringStatusSentinel
)

func (t MonitoringStatus) IsValid() bool {
	return t > MonitoringStatusUnknown && t < monitoringStatusSentinel
}

func (t MonitoringStatus) String() string {
	switch t {
	case MonitoringStatusAll:
		return "ALL"
	case MonitoringStatusFuture:
		return "FUTURE"
	case MonitoringStatusMissing:
		return "MISSING"
	case MonitoringStatusExisting:
		return "EXISTING"
	case MonitoringStatusFirstSeason:
		return "FIRST_SEASON"
	case MonitoringStatusLatestSeason:
		return "LATEST_SEASON"
	case MonitoringStatusNone:
		return "NONE"
	default:
		return "UNKNOWN"
	}
}
