package statusbar

import (
	"context"
	"time"
)

type Updater struct {
	ctx   context.Context
	store *Store
}

func NewUpdater(ctx context.Context, store *Store) *Updater {
	return &Updater{ctx: ctx, store: store}
}

func (u *Updater) Register(key string, interval time.Duration, fn func() string) {
	go func() {
		u.store.Set(key, fn())
		for {
			select {
			case <-time.After(interval):
				u.store.Set(key, fn())
			case <-u.ctx.Done():
				return
			}
		}
	}()
}
