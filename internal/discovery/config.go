package discovery

import (
	"time"

	"github.com/expr-lang/expr/vm"
)

type Config struct {
	SearchSuffix     string
	ReleaseGroups    []string
	Qualitites       []string
	Category         string
	RenameTorrent    bool
	RenameFormat     *vm.Program
	DownloadPath     string
	CreateShowFolder bool
	PollFrequency    time.Duration
}
