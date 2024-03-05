package notifier_test

import (
	"github.com/sealbro/go-feed-me/pkg/notifier"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBatching(t *testing.T) {
	testCases := []struct {
		name         string
		items        []int
		maxTimeout   time.Duration
		maxItems     int
		expectGroups int
		delay        time.Duration
	}{
		{
			name:         "by max items with max 3",
			items:        []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			maxTimeout:   1 * time.Minute,
			maxItems:     3,
			expectGroups: 4,
		},
		{
			name:         "by max items with max 1",
			items:        []int{1, 2, 3, 4, 5},
			maxTimeout:   1 * time.Second,
			maxItems:     1,
			expectGroups: 5,
		},
		{
			name:         "by time",
			items:        []int{1, 2, 3},
			maxTimeout:   1 * time.Microsecond,
			maxItems:     5,
			expectGroups: 3,
			delay:        1 * time.Millisecond,
		},
	}

	for i := range testCases {
		testCase := testCases[i]

		t.Run(testCase.name, func(t *testing.T) {
			RunTestCase(t, testCase)
		})
	}
}

func RunTestCase(t *testing.T, testCase struct {
	name         string
	items        []int
	maxTimeout   time.Duration
	maxItems     int
	expectGroups int
	delay        time.Duration
}) {
	t.Parallel()
	values := make(chan int)
	go func() {
		for _, item := range testCase.items {
			values <- item
			time.Sleep(testCase.delay)
		}

		close(values)
	}()

	process := notifier.SplitByBatchProcess(values, testCase.maxItems, testCase.maxTimeout)

	i := 0
	groups := 0
	for grouped := range process {
		for _, g := range grouped {
			assert.Equal(t, testCase.items[i], g, "grouped item should be equal to original item")
			i++
		}
		groups++
	}
	assert.Equal(t, len(testCase.items), i, "number of items should be equal to original")
	assert.Equal(t, testCase.expectGroups, groups, "number of groups should be equal to expected")
}
