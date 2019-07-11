// Copyright 2014 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"context"
	"errors"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/prometheus/prometheus/tsdb/chunks"
)

// The errors exposed.
var (
	ErrNotFound                    = errors.New("not found")
	ErrOutOfOrderSample            = errors.New("out of order sample")
	ErrDuplicateSampleForTimestamp = errors.New("duplicate sample for timestamp")
	ErrOutOfBounds                 = errors.New("out of bounds")
)

// Storage ingests and manages samples, along with various indexes. All methods
// are goroutine-safe. Storage implements storage.SampleAppender.
type Storage interface {
	Queryable

	// StartTime returns the oldest timestamp stored in the storage.
	StartTime() (int64, error)

	// Appender returns a new appender against the storage.
	Appender() (Appender, error)

	// Close closes the storage and all its underlying resources.
	Close() error
}

// A Queryable handles queries against a storage.
type Queryable interface {
	// Querier returns a new Querier on the storage.
	Querier(ctx context.Context, mint, maxt int64) (Querier, error)
}

// Querier provides querying access over time series data of a fixed
// time range.
type Querier interface {
	// Select returns a set of series that matches the given label matchers.
	Select(...*labels.Matcher) (SeriesSet, error)

	// SelectSorted returns a sorted set of series that matches the given label matcher.
	SelectSorted(...*labels.Matcher) (SeriesSet, error)

	// LabelValues returns all potential values for a label name.
	// It is not safe to use the strings beyond the lifefime of the querier.
	LabelValues(string) ([]string, error)

	// LabelNames returns all the unique label names present in the block in sorted order.
	LabelNames() ([]string, error)

	// Close releases the resources of the Querier.
	Close() error
}


// SelectParams specifies parameters passed to data selections.
type SelectParams struct {
	Start int64 // Start time in milliseconds for this select.
	End   int64 // End time in milliseconds for this select.

	Step int64  // Query step size in milliseconds.
	Func string // String representation of surrounding function or aggregation.

	Grouping []string // List of label names used in aggregation.
	By       bool     // Indicate whether it is without or by.
	Range    int64    // Range vector selector range in milliseconds.
}

// QueryableFunc is an adapter to allow the use of ordinary functions as
// Queryables. It follows the idea of http.HandlerFunc.
type QueryableFunc func(ctx context.Context, mint, maxt int64) (Querier, error)

// Querier calls f() with the given parameters.
func (f QueryableFunc) Querier(ctx context.Context, mint, maxt int64) (Querier, error) {
	return f(ctx, mint, maxt)
}

// Appender provides batched appends against a storage.
type Appender interface {
	Add(l labels.Labels, t int64, v float64) (uint64, error)

	AddFast(l labels.Labels, ref uint64, t int64, v float64) error

	// Commit submits the collected samples and purges the batch.
	Commit() error

	Rollback() error
}

// SeriesSet contains a set of series.
type SeriesSet interface {
	Next() bool
	At() Series
	Err() error
}

// ChunkSeriesSet exposes the chunks and intervals of a series instead of the
// actual series itself.
type ChunkSeriesSet interface {
	Next() bool
	At() ChunkSeries
	Err() error
}

// Series exposes a single time series and allows to iterate over samples as well chunks.
type Series interface {
	// Labels returns the complete set of labels identifying the series.
	Labels() labels.Labels

	// Iterator returns a new iterator of the data of the series.
	Iterator() chunkenc.Iterator
}

// Series exposes a single time series and allows to iterate over samples as well chunks.
type ChunkSeries interface {
	// Labels returns the complete set of labels identifying the series.
	Labels() labels.Labels

	// ChunkIterator returns a new iterator that iterates over non-overlapping chunks of the series.
	Iterator() ChunkIterator
}

// ChunkIterator iterates over the chunk of a time series.
type ChunkIterator interface {
	//// Seek advances the iterator forward to the given timestamp.
	//// It advances to the chunk with min time at t or first chunk with min time after t.
	//Seek(t int64) bool
	// At returns the current meta.
	At() chunks.Meta
	// Next advances the iterator by one.
	Next() bool
	// Err returns optional error if Next is false.
	Err() error
}

// SeriesIterator iterates over the data of a time series.
type SeriesIterator interface {
	// Seek advances the iterator forward to the value at or after
	// the given timestamp.
	Seek(t int64) bool
	// At returns the current timestamp/value pair.
	At() (t int64, v float64)
	// Next advances the iterator by one.
	Next() bool
	// Err returns the current error.
	Err() error
}

type Warnings []error
