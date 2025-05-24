package runner

import (
	"context"
	"errors"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func ptrTo[T any](t T) *T {
	return &t
}

func messages(errs []error) []string {
	errMsgs := make([]string, 0, len(errs))
	for _, e := range errs {
		errMsgs = append(errMsgs, e.Error())
	}
	return errMsgs
}

func TestRunner(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name           string
		maxConcurrency uint
		totalJobs      uint32
		wantErrCount   uint32
	}{
		{
			name:           "no_max_concurrency_all_ok",
			maxConcurrency: 0,
			totalJobs:      64,
		},
		{
			name:           "no_max_concurrency_some_errs",
			maxConcurrency: 0,
			totalJobs:      64,
			wantErrCount:   32,
		},
		{
			name:           "no_max_concurrency_all_errs",
			maxConcurrency: 0,
			totalJobs:      64,
			wantErrCount:   64,
		},
		{
			name:           "max_concurrency_1_all_ok",
			maxConcurrency: 1,
			totalJobs:      64,
		},
		{
			name:           "max_concurrency_1_some_errs",
			maxConcurrency: 1,
			totalJobs:      64,
			wantErrCount:   16,
		},
		{
			name:           "max_concurrency_1_all_errs",
			maxConcurrency: 1,
			totalJobs:      64,
			wantErrCount:   64,
		},
		{
			name:           "max_concurrency_16_all_ok",
			maxConcurrency: 16,
			totalJobs:      64,
		},
		{
			name:           "max_concurrency_16_some_errs",
			maxConcurrency: 16,
			totalJobs:      64,
			wantErrCount:   16,
		},
		{
			name:           "max_concurrency_16_all_errs",
			maxConcurrency: 16,
			totalJobs:      64,
			wantErrCount:   64,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner := New(context.TODO(), WithLimit(tc.maxConcurrency))
			successCount := atomic.Uint32{}
			remainingErrCount := tc.wantErrCount

			for range tc.totalJobs {
				var jobResult error
				if remainingErrCount > 0 {
					jobResult = errors.New("test error")
					remainingErrCount--
				}

				// Make goroutines more likely to last long enough to overlap:
				randGenSrc := rand.NewSource(time.Now().UnixNano())
				randGen := rand.New(randGenSrc)
				sleepDuration := time.Duration(randGen.Intn(50)) * time.Millisecond

				runner.Go(func() error {
					time.Sleep(sleepDuration)
					if jobResult == nil {
						successCount.Add(1)
					}
					return jobResult
				})
			}
			errs := runner.Wait()

			gotErrCount := uint32(len(errs))
			if gotErrCount != tc.wantErrCount {
				t.Errorf("Wait() did not return the expected number of errors, got %d, want %d", gotErrCount, tc.wantErrCount)
			}
			wantSuccessCount := tc.totalJobs - tc.wantErrCount
			if successCount.Load() != wantSuccessCount {
				t.Errorf("Wait() did not return the expected number of successes, got %d, want %d", successCount.Load(), wantSuccessCount)
			}
		})
	}
}

func TestRunnerOption_WithLimit(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		limit *uint
	}{
		{
			name: "limit_not_specified",
		},
		{
			name:  "limit_0",
			limit: ptrTo[uint](0),
		},
		{
			name:  "limit_1",
			limit: ptrTo[uint](1),
		},
		{
			name:  "limit_16",
			limit: ptrTo[uint](16),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var opts []Option
			var wantCapacity uint
			if tc.limit != nil {
				wantCapacity = *tc.limit
				opts = append(opts, WithLimit(wantCapacity))
			}
			runner := New(context.Background(), opts...)

			if runner.hasLimit() && wantCapacity == 0 {
				t.Errorf("hasConcurrencyLimit() did not return the true when capacity should have been 0")
			}
			if uint(cap(runner.sem)) != wantCapacity {
				t.Errorf("sem capacity was %d, want %d", cap(runner.sem), wantCapacity)
			}
		})
	}
}

func TestRunnerCancelOnFailure(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		totalJobs    int
		failAtIndex  *int
		wantErrCount int
	}{
		{
			name:      "all_tasks_ok",
			totalJobs: 8,
		},
		{
			name:         "first_task_fails",
			totalJobs:    8,
			failAtIndex:  ptrTo[int](0),
			wantErrCount: 8,
		},
		{
			name:         "second_task_fails",
			totalJobs:    8,
			failAtIndex:  ptrTo[int](1),
			wantErrCount: 7,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner := New(context.TODO(), WithCancelOnFailure())

			for i := range tc.totalJobs {
				var jobResult error
				if tc.failAtIndex != nil && *tc.failAtIndex == i {
					jobResult = errors.New("test error")
				}
				runner.Go(func() error {
					return jobResult
				})
				// Wait for the task to finish before queueing up the next one; this essentially
				// makes them run sequentially, making it easier for us to reason about the sequence
				// of events.
				_ = runner.Wait()
			}
			errs := runner.Wait()
			gotErrCount := len(errs)

			// In sequential mode, all failure causes after the first should be context.Canceled.
			wantCanceledReasonCount := tc.wantErrCount - 1
			if tc.wantErrCount == 0 {
				wantCanceledReasonCount = 0
			}
			gotCanceledReasonCount := 0
			for _, err := range errs {
				if errors.Is(err, context.Canceled) {
					gotCanceledReasonCount++
				}
			}

			if gotErrCount != tc.wantErrCount {
				t.Errorf("Wait() did not return the expected number of errors, got %d, want %d", gotErrCount, tc.wantErrCount)
			}
			if gotCanceledReasonCount != wantCanceledReasonCount {
				t.Errorf("Wait() did not return the expected number of context.Canceled errors, got %d, want %d; errs: %#v", gotCanceledReasonCount, wantCanceledReasonCount, messages(errs))
			}
		})
	}
}

func TestRunnerCancelOnParentContextRespected(t *testing.T) {
	for _, tc := range []struct {
		name              string
		cancelOnFailure   bool
		totalJobs         int
		cancelBeforeIndex int
		wantErrCount      int
	}{
		{
			name:              "NoCancelOnFailure_parent_canceled_before_first_task",
			totalJobs:         8,
			cancelBeforeIndex: 0,
			wantErrCount:      8,
		},
		{
			name:              "CancelOnFailure_parent_canceled_before_first_task",
			cancelOnFailure:   true,
			totalJobs:         8,
			cancelBeforeIndex: 0,
			wantErrCount:      8,
		},
		{
			name:              "NoCancelOnFailure_parent_canceled_before_third_task",
			totalJobs:         8,
			cancelBeforeIndex: 2,
			wantErrCount:      6,
		},
		{
			name:              "CancelOnFailure_parent_canceled_before_third_task",
			cancelOnFailure:   true,
			totalJobs:         8,
			cancelBeforeIndex: 2,
			wantErrCount:      6,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Regardless of whether the WithContinueOnFailure option is specified, the runner
			// should still respect the cancellation of the context provided at initialization.
			parentCtx, parentCancel := context.WithCancel(context.Background())
			var opts []Option
			if tc.cancelOnFailure {
				opts = append(opts, WithCancelOnFailure())
			}
			runner := New(parentCtx, opts...)

			for i := range tc.totalJobs {
				if tc.cancelBeforeIndex == i {
					// To ensure all tasks up until this point have finished and succeeded, we must
					// synchronize via the runner's waitgroup before canceling the context.
					_ = runner.Wait()
					parentCancel()
				}
				runner.Go(func() error {
					return nil
				})
			}
			errs := runner.Wait()
			gotErrCount := len(errs)

			if gotErrCount != tc.wantErrCount {
				t.Errorf("Wait() did not return the expected number of errors, got %d, want %d; errs: %#v", gotErrCount, tc.wantErrCount, messages(errs))
			}
			_ = parentCancel // Suppress false-positive "lostcancel" linter warning.
		})
	}
}

