package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/collections"
)

func NewMonitoringStatus(from sqlcgen.MonitoringStatus) collections.MonitoringStatus {
	switch from {
	case sqlcgen.MonitoringStatusAll:
		return collections.MonitoringStatusAll
	case sqlcgen.MonitoringStatusExisting:
		return collections.MonitoringStatusExisting
	case sqlcgen.MonitoringStatusFirstSeason:
		return collections.MonitoringStatusFirstSeason
	case sqlcgen.MonitoringStatusFuture:
		return collections.MonitoringStatusFuture
	case sqlcgen.MonitoringStatusLatestSeason:
		return collections.MonitoringStatusLatestSeason
	case sqlcgen.MonitoringStatusMissing:
		return collections.MonitoringStatusMissing
	case sqlcgen.MonitoringStatusNone:
		return collections.MonitoringStatusNone
	default:
		return collections.MonitoringStatusUnknown
	}
}

func NewMonitoringStatusModel(from collections.MonitoringStatus) sqlcgen.MonitoringStatus {
	switch from {
	case collections.MonitoringStatusAll:
		return sqlcgen.MonitoringStatusAll
	case collections.MonitoringStatusExisting:
		return sqlcgen.MonitoringStatusExisting
	case collections.MonitoringStatusFirstSeason:
		return sqlcgen.MonitoringStatusFirstSeason
	case collections.MonitoringStatusFuture:
		return sqlcgen.MonitoringStatusFuture
	case collections.MonitoringStatusLatestSeason:
		return sqlcgen.MonitoringStatusLatestSeason
	case collections.MonitoringStatusMissing:
		return sqlcgen.MonitoringStatusMissing
	case collections.MonitoringStatusNone:
		return sqlcgen.MonitoringStatusNone
	default:
		return sqlcgen.MonitoringStatusUnknown
	}
}
