package domain

type ReleaseStatus uint

const (
	ReleaseStatusUnknown ReleaseStatus = iota
	ReleaseStatusPending
	ReleaseStatusGrabbed
	ReleaseStatusRejected
	ReleaseStatusImported
	ReleaseStatusFailed
)

func (r ReleaseStatus) String() string {
	switch r {
	case ReleaseStatusPending:
		return "PENDING"
	case ReleaseStatusGrabbed:
		return "GRABBED"
	case ReleaseStatusRejected:
		return "REJECTED"
	case ReleaseStatusImported:
		return "IMPORTED"
	case ReleaseStatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

func (r ReleaseStatus) IsValid() bool {
	return r > ReleaseStatusUnknown && r <= ReleaseStatusFailed
}
