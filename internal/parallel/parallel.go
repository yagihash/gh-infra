package parallel

import "sync"

// Map processes items concurrently and returns results in input order.
// concurrency controls the maximum number of goroutines running at once.
// If concurrency <= 0, all items are processed without a limit.
func Map[T, R any](items []T, concurrency int, fn func(int, T) R) []R {
	results := make([]R, len(items))
	if len(items) == 0 {
		return results
	}
	if concurrency <= 0 {
		concurrency = len(items)
	}

	ch := make(chan int, len(items))

	var wg sync.WaitGroup
	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range ch {
				results[i] = fn(i, items[i])
			}
		}()
	}

	for i := range items {
		ch <- i
	}
	close(ch)

	wg.Wait()
	return results
}
