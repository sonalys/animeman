package discovery

import (
	"reflect"
	"testing"

	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/stretchr/testify/require"
)

func Test_filterEpisodes(t *testing.T) {
	type args struct {
		list      []TaggedNyaa
		latestTag string
	}
	tests := []struct {
		name string
		args args
		want []TaggedNyaa
	}{
		{
			name: "empty",
			args: args{},
			want: []TaggedNyaa{},
		},
		{
			name: "no tag",
			args: args{
				latestTag: "",
				list: []TaggedNyaa{
					{seasonEpisodeTag: "!Show3 S03E01"},
					{seasonEpisodeTag: "!Show3 S03E02"},
				},
			},
			want: []TaggedNyaa{
				{seasonEpisodeTag: "!Show3 S03E01"},
				{seasonEpisodeTag: "!Show3 S03E02"},
			},
		},
		{
			name: "tag",
			args: args{
				latestTag: "!Show3 S03E01",
				list: []TaggedNyaa{
					{seasonEpisodeTag: "!Show3 S03E01"},
					{seasonEpisodeTag: "!Show3 S03E02"},
				},
			},
			want: []TaggedNyaa{
				{seasonEpisodeTag: "!Show3 S03E02"},
			},
		},
		{
			name: "season batch",
			args: args{
				latestTag: "!Show3 S03",
				list: []TaggedNyaa{
					{seasonEpisodeTag: "!Show3 S03E01"},
					{seasonEpisodeTag: "!Show3 S03E02"},
				},
			},
			want: []TaggedNyaa{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterEpisodes(tt.args.list, tt.args.latestTag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterEpisodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildTaggedNyaaList(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := buildTaggedNyaaList([]nyaa.Entry{})
		require.Empty(t, got)
	})
	t.Run("sort by tag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
			{Title: "Show3: S03"},
		}
		got := buildTaggedNyaaList(input)
		require.Len(t, got, len(input))
		latestTag := got[len(got)-1].seasonEpisodeTag
		for i := range got {
			require.True(t, compareTags(got[i].seasonEpisodeTag, latestTag) <= 0)
		}
	})
}
