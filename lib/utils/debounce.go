/*
Copyright 2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"context"
	"time"
)

// Debouncer wraps a channel to limit the it will fire at.
type Debouncer interface {
	Debounce(ctx context.Context, c <-chan struct{}) <-chan struct{}
}

// NewDebouncer returns a new Debouncer instance.
func NewDebouncer(initialBackoff, maxBackoff time.Duration) Debouncer {
	return &debouncer{
		initialBackoff: initialBackoff,
		maxBackoff:     maxBackoff,
	}
}

// Debouncer wraps a channel to limit the rate it will fire at. The backoff is
// doubled if the channel is received on before the current backoff is
// complete. Otherwise the backoff is reset to the initial value.
type debouncer struct {
	initialBackoff time.Duration
	maxBackoff     time.Duration

	c         <-chan struct{}
	debounced chan struct{}

	currBackoff time.Duration
	ticker      *time.Ticker
}

func (d *debouncer) Debounce(ctx context.Context, c <-chan struct{}) <-chan struct{} {
	d.c = c
	d.debounced = make(chan struct{})

	if d.currBackoff == 0 {
		d.currBackoff = d.initialBackoff
	}

	d.ticker = time.NewTicker(d.currBackoff)

	go func() {
		defer d.ticker.Stop()
		defer close(d.debounced)

		for {
			ok := d.debounceSend(ctx)
			if !ok {
				return
			}
		}
	}()

	return d.debounced
}

func (d *debouncer) send() bool {
	select {
	case d.debounced <- struct{}{}:
		return true
	default:
		return false
	}
}

func (d *debouncer) debounceSend(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case _, ok := <-d.c:
		if !ok {
			return false
		}
	}

	select {
	case <-d.ticker.C:
		d.send()
		d.currBackoff = d.initialBackoff
		d.ticker.Reset(d.initialBackoff)
	default:
		d.backoffSend(ctx)
	}

	return true
}

// backoffSend waits for the ticker before sending and increases the
// exponential backoff if the send went through.
func (d *debouncer) backoffSend(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-d.ticker.C:
		ok := d.send()
		if !ok {
			d.currBackoff = d.initialBackoff
			d.ticker.Reset(d.initialBackoff)
		}

		d.currBackoff = d.initialBackoff * 2
		if d.currBackoff > d.maxBackoff {
			d.currBackoff = d.maxBackoff
		}
	}
}
