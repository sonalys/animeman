package utils

import (
	"maps"
	"slices"
)

func Map[T1, T2 any](in []T1, f func(T1) T2) []T2 {
	out := make([]T2, 0, len(in))
	for i := range in {
		out = append(out, f(in[i]))
	}
	return out
}

func Filter[T any](in []T, filters ...func(T) bool) []T {
	out := make([]T, 0, len(in))
outer:
	for i := range in {
		for _, filter := range filters {
			if !filter(in[i]) {
				continue outer
			}
		}
		out = append(out, in[i])
	}
	return out
}

func Deduplicate[T comparable](from []T) []T {
	set := make(map[T]struct{}, len(from))

	for i := range from {
		set[from[i]] = struct{}{}
	}

	return slices.Collect(maps.Keys(set))
}
