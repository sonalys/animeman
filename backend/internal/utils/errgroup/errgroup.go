package errgroup

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type Group struct{ group *errgroup.Group }

func WithContext(ctx context.Context) (*Group, context.Context) {
	errgrp, errctx := errgroup.WithContext(ctx)
	return &Group{group: errgrp}, errctx
}

func (g *Group) Go(f func() error) {
	g.group.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic recovered: %v", r)
			}
		}()

		return f()
	})
}

func (g *Group) SetLimit(n int) {
	g.group.SetLimit(n)
}

func (g *Group) TryGo(f func() error) bool {
	return g.group.TryGo(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic recovered: %v", r)
			}
		}()

		return f()
	})
}

func (g *Group) Wait() error {
	return g.group.Wait()
}
