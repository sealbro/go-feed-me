package notifier

import (
	"time"
)

// SplitByBatchProcess batch processor that will split a channel into batches of a maximum size or after a maximum timeout.
// main idea from https://elliotchance.medium.com/batch-a-channel-by-size-or-time-in-go-92fa3098f65
func SplitByBatchProcess[TItem any](values <-chan TItem, maxItems int, maxTimeout time.Duration) <-chan []TItem {
	batches := make(chan []TItem)

	go func() {
		defer close(batches)
		for keepGoing := true; keepGoing; {
			var batch []TItem
			batch, keepGoing = getBatch(values, maxItems, maxTimeout)

			if len(batch) > 0 {
				batches <- batch
			}
		}
	}()

	return batches
}

func getBatch[TItem any](values <-chan TItem, maxItems int, maxTimeout time.Duration) ([]TItem, bool) {
	keepGoing := true

	var batch []TItem
	expire := time.After(maxTimeout)
	for {
		refreshBatch := false

		select {
		case value, ok := <-values:
			if !ok {
				keepGoing = false
				refreshBatch = true
				break
			}

			batch = append(batch, value)
			if len(batch) >= maxItems {
				refreshBatch = true
				break
			}
		case <-expire:
			refreshBatch = true
			break
		}

		if refreshBatch {
			break
		}
	}

	return batch, keepGoing
}
