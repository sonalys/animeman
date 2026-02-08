package discovery

import (
	"testing"

	"github.com/sonalys/animeman/internal/utils"
	"github.com/stretchr/testify/require"
)

func Test_matchTitle(t *testing.T) {
	testCases := []struct {
		name string
		a, b string
		want bool
	}{
		{
			name: "do not match different titles",
			a:    "anime 1",
			b:    "anime 2",
			want: false,
		},
		{
			name: "match prefix",
			a:    "anime 1: subtitle",
			b:    "anime 1",
			want: true,
		},
		{
			name: "ignore special characters",
			a:    "anime 1/2",
			b:    "anime 1-2",
			want: true,
		},
		{
			name: "ignore space",
			a:    "anime1/2",
			b:    "anime 1-2",
			want: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			match := utils.MatchPrefixFlexible(tC.a, tC.b, ignoreCharset)
			require.Equal(t, tC.want, match)
		})
	}
}
