package collections

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
		return "pending"
	case ReleaseStatusGrabbed:
		return "grabbed"
	case ReleaseStatusRejected:
		return "rejected"
	case ReleaseStatusImported:
		return "imported"
	case ReleaseStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

func (r ReleaseStatus) IsValid() bool {
	return r > ReleaseStatusUnknown && r <= ReleaseStatusFailed
}
