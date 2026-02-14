package transfer

type Status int

const (
	StatusUnknown Status = iota
	StatusPending
	StatusDownloading
	StatusImporting
	StatusCompleted
	StatusFailed
	statusSentinel
)

func (s Status) IsValid() bool {
	return s > StatusUnknown && s < statusSentinel
}

var statusNames = map[Status]string{
	StatusPending:     "pending",
	StatusDownloading: "downloading",
	StatusImporting:   "importing",
	StatusCompleted:   "completed",
	StatusFailed:      "failed",
}

// String implements the fmt.Stringer interface.
func (s Status) String() string {
	if name, ok := statusNames[s]; ok {
		return name
	}
	return "unknown"
}
