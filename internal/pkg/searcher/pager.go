package searcher

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/pkg/parser"
	"github.com/sonalys/animeman/internal/pkg/sliceutils"
	"github.com/sonalys/animeman/internal/pkg/tags"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

type (
	FetchFunc  = func(ctx context.Context, offset int) ([]torrentsource.Torrent, error)
	FilterFunc = func(ignoreCounter func(string)) func(torrent torrentsource.Torrent) bool

	Searcher struct {
		pageSize          int
		fetch             FetchFunc
		additionalFilters []FilterFunc
	}
)

func New(pageSize int, fetch FetchFunc, additionalFilters ...FilterFunc) Searcher {
	return Searcher{
		pageSize:          pageSize,
		fetch:             fetch,
		additionalFilters: additionalFilters,
	}
}

// Search will return all relevant torrent results for the anime list entry and search options.
// It will prioritize and paginate based on release group and video preferences, as well for the latest present tag.
// It returns a sorted list of torrent candidates that are newer than the provided latest tag.
func (p Searcher) Search(
	ctx context.Context,
	entry animelist.Entry,
	opts torrentsource.SearchOptions,
) ([]torrentsource.Torrent, error) {
	torrents := make([]torrentsource.Torrent, 0, p.pageSize)
	offset := 0

	filterStats := newFilterStats()

	unregisteredFilters := []FilterFunc{
		minSeeders(1),
		matchStartDate(entry),
		matchEpisodeCount(entry, opts.Sources),
		matchSources(opts.Sources),
		newerEpisode(opts.LatestTag),
	}
	unregisteredFilters = append(unregisteredFilters, p.additionalFilters...)
	registeredFilters := sliceutils.Map(
		unregisteredFilters,
		func(f FilterFunc) func(torrentsource.Torrent) bool {
			return f(filterStats.Inc)
		},
	)

	for {
		items, err := p.fetch(ctx, offset)
		if err != nil {
			return nil, err
		}

		log.
			Ctx(ctx).
			Trace().
			Any("torrents", items).
			Msg("received pagination result")

		if len(items) == 0 {
			break
		}

		offset += len(items)

		filtered := sliceutils.Filter(items, registeredFilters...)
		torrents = append(torrents, filtered...)

		if len(items) < p.pageSize || !shouldPaginate(filtered, opts.LatestTag) {
			break
		}
	}

	torrents = prioritize(entry, torrents, opts)

	log.
		Ctx(ctx).
		Debug().
		Int("results", len(torrents)).
		Any("ignored", filterStats).
		Msg("search results")

	return torrents, nil
}

// shouldPaginate reports whether the source may have more results after this
// page. nekoBT returns newest results first, so paginate while even the
// smallest tag found remains newer than the latest downloaded tag.
func shouldPaginate(items []torrentsource.Torrent, latestTag tags.Tag) bool {
	if latestTag.IsZero() {
		// Nothing downloaded yet: the first page already has everything.
		return false
	}

	if len(items) == 0 {
		return false
	}

	var smallest tags.Tag

	for _, it := range items {
		tag := parser.Parse(it.Title, 1, nil).Tag
		if smallest.IsZero() || tag.Compare(smallest) < 0 {
			smallest = tag
		}
	}

	return smallest.Compare(latestTag) > 0
}
