package plextrac

import (
	"context"
	"fmt"
	"net/url"
)

// ListOpts is the common set of query parameters accepted by list endpoints.
type ListOpts struct {
	Page    int
	PerPage int
	Search  string
	Extra   url.Values
}

func (o ListOpts) query() url.Values {
	q := url.Values{}
	for k, vs := range o.Extra {
		q[k] = append([]string(nil), vs...)
	}
	if o.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", o.Page))
	}
	if o.PerPage > 0 {
		q.Set("per_page", fmt.Sprintf("%d", o.PerPage))
	}
	if o.Search != "" {
		q.Set("search", o.Search)
	}
	return q
}

// Iter is a lazy paginator over a resource list. Next fetches the next page
// when the local buffer is exhausted. Consumers must check Err after Next
// returns false.
type Iter[T any] struct {
	ctx     context.Context
	fetch   func(ctx context.Context, page int) ([]T, bool, error)
	buf     []T
	idx     int
	page    int
	hasMore bool
	err     error
}

// NewIter constructs an iterator from a page-fetch closure.
func newIter[T any](ctx context.Context, fetch func(ctx context.Context, page int) ([]T, bool, error)) *Iter[T] {
	return &Iter[T]{ctx: ctx, fetch: fetch, page: 1, hasMore: true}
}

// Next advances the iterator. Returns false when exhausted or on error.
func (it *Iter[T]) Next(ctx context.Context) bool {
	if it.err != nil {
		return false
	}
	if it.idx < len(it.buf) {
		it.idx++
		return true
	}
	if !it.hasMore {
		return false
	}
	items, more, err := it.fetch(ctx, it.page)
	if err != nil {
		it.err = err
		return false
	}
	it.buf = items
	it.idx = 0
	it.hasMore = more
	it.page++
	if len(it.buf) == 0 {
		return false
	}
	it.idx = 1
	return true
}

// Value returns the current page element.
func (it *Iter[T]) Value() T {
	if it.idx == 0 || it.idx > len(it.buf) {
		var zero T
		return zero
	}
	return it.buf[it.idx-1]
}

// Err returns any error encountered while fetching pages.
func (it *Iter[T]) Err() error { return it.err }

// All materialises the full iterator into a slice. Convenient for small
// result sets; avoid on large reports.
func (it *Iter[T]) All(ctx context.Context) ([]T, error) {
	var out []T
	for it.Next(ctx) {
		out = append(out, it.Value())
	}
	return out, it.Err()
}
