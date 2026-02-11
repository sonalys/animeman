package discovery

import (
	"testing"

	"github.com/sonalys/animeman/internal/domain"
	"github.com/sonalys/animeman/internal/utils/tags"
	"github.com/stretchr/testify/require"
)

func Test_getLatestTag(t *testing.T) {
	type args struct {
		torrents []domain.Torrent
	}
	tests := []struct {
		name string
		args args
		want tags.Tag
	}{
		{
			name: "batch and season",
			args: args{
				torrents: []domain.Torrent{
					{Tags: []string{"S1E1~13"}},
					{Tags: []string{"S1E2"}},
					{Tags: []string{"S1E3"}},
				},
			},
			want: tags.Tag{
				Seasons:  []int{1},
				Episodes: []float64{1, 13},
			},
		},
		{
			name: "same season half episode",
			args: args{
				torrents: []domain.Torrent{
					{Tags: []string{"ore dake level up na ken S1E7"}},
					{Tags: []string{"solo leveling S1E7.5"}},
				},
			},
			want: tags.Tag{
				Seasons:  []int{1},
				Episodes: []float64{7.5},
			},
		},
		{
			name: "batch and same season",
			args: args{
				torrents: []domain.Torrent{
					{Tags: []string{"S3"}},
					{Tags: []string{"S3E2"}},
				},
			},
			want: tags.Tag{
				Seasons: []int{3},
			},
		},
		{
			name: "empty",
			want: tags.Tag{},
		},
		{
			name: "one tag",
			args: args{
				torrents: []domain.Torrent{
					{Tags: []string{"S01"}},
				},
			},
			want: tags.Tag{
				Seasons: []int{1},
			},
		},
		{
			name: "same season",
			args: args{
				torrents: []domain.Torrent{
					{Tags: []string{"S1E1"}},
					{Tags: []string{"S1E2"}},
					{Tags: []string{"S1E3"}},
				},
			},
			want: tags.Tag{
				Seasons:  []int{1},
				Episodes: []float64{3},
			},
		},
		{
			name: "different seasons",
			args: args{
				torrents: []domain.Torrent{
					{Tags: []string{"S3E1"}},
					{Tags: []string{"S2E2"}},
					{Tags: []string{"S1E3"}},
				},
			},
			want: tags.Tag{
				Seasons:  []int{3},
				Episodes: []float64{1},
			},
		},
		{
			name: "batch and season",
			args: args{
				torrents: []domain.Torrent{
					{Tags: []string{"S3"}},
					{Tags: []string{"S2E2"}},
					{Tags: []string{"S1E3"}},
				},
			},
			want: tags.Tag{
				Seasons: []int{3},
			},
		},
		{
			name: "batches of different seasons",
			args: args{
				torrents: []domain.Torrent{
					{Tags: []string{"S3"}},
					{Tags: []string{"S4"}},
				},
			},
			want: tags.Tag{
				Seasons: []int{4},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLatestTag(tt.args.torrents); tagCompare(got, tt.want) != 0 {
				t.Errorf("getLatestTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mergeBatchEpisodes(t *testing.T) {
	type args struct {
		tag string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{},
			want: "",
		},
		{
			name: "ok",
			args: args{
				tag: "S03E1~12",
			},
			want: "S03E12",
		},
		{
			name: "no batch",
			args: args{
				tag: "S03E1",
			},
			want: "S03E1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tagMergeBatchEpisodes(tt.args.tag); got != tt.want {
				t.Errorf("mergeBatchEpisodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tagCompare(t *testing.T) {
	t.Run("same tag", func(t *testing.T) {
		tag := tags.Tag{
			Seasons:  []int{3},
			Episodes: []float64{2},
		}
		require.Zero(t, tagCompare(tag, tag))
	})

	t.Run("first episode and zero tag", func(t *testing.T) {
		tag := tags.Tag{
			Seasons:  []int{1},
			Episodes: []float64{1},
		}
		require.Greater(t, tagCompare(tag, tags.Tag{}), 0)
	})

	t.Run("batch different season", func(t *testing.T) {
		tagA := tags.Tag{
			Seasons: []int{2},
		}

		tagB := tags.Tag{
			Seasons: []int{3},
		}

		require.Equal(t, tagCompare(tagA, tagB), -1)
	})

	t.Run("batch and single", func(t *testing.T) {
		tagA := tags.Tag{
			Seasons: []int{2},
		}

		tagB := tags.Tag{
			Seasons:  []int{2},
			Episodes: []float64{1},
		}

		require.Equal(t, tagCompare(tagA, tagB), 1)
	})

	t.Run("batch and zero", func(t *testing.T) {
		tagA := tags.Tag{
			Seasons: []int{1},
		}

		require.Equal(t, tagCompare(tagA, tags.Zero), 1)
	})
}
