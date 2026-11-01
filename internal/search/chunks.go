package search

import (
	"slices"
	"sync"
)

// chunkLines is how many lines one goroutine searches. Shorter texts are
// searched in one go.
var chunkLines = 10_000

// inChunks runs find on pieces of the rows from 0 up to count at the same
// time and joins the results in the order of the rows.
func inChunks[T any](count int, find func(from, to int) []T) []T {
	pieces := (count + chunkLines - 1) / chunkLines
	if pieces <= 1 {
		return find(0, count)
	}
	results := make([][]T, pieces)
	var wg sync.WaitGroup
	for i := range results {
		wg.Go(func() {
			results[i] = find(i*chunkLines, min((i+1)*chunkLines, count))
		})
	}
	wg.Wait()
	return slices.Concat(results...)
}
