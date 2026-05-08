package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestScheduler_RunCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := callRunByCtx(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestScheduler_RunTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), -time.Second)
	defer cancel()

	err := callRunByCtx(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestScheduler_runOnceSuccessful(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetchedItems := []int{42}
	dispatchedItems := make(chan int, 1)

	s := New[int]()

	err := s.runOnce(ctx,
		func(ctx context.Context) ([]int, error) {
			return fetchedItems, nil
		},
		func(ctx context.Context, item int) error {
			dispatchedItems <- item
			cancel()
			return nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, 42, <-dispatchedItems)
}

func TestScheduler_runOnceCanceledBeforeNextItem(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetchedItems := []int{1, 2}
	dispatchedItems := make(chan int, 1)
	s := New[int]()

	err := s.runOnce(ctx,
		func(ctx context.Context) ([]int, error) {
			return fetchedItems, nil
		},
		func(ctx context.Context, item int) error {
			if item == 1 {
				dispatchedItems <- item
				cancel()
			}

			return nil
		},
	)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, <-dispatchedItems)
}

func TestScheduler_runOnceTimeoutBeforeNextItem(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	fetchedItems := []int{1, 2}
	dispatchedItems := make(chan int, 1)
	s := New[int]()

	err := s.runOnce(
		ctx,
		func(ctx context.Context) ([]int, error) {
			return fetchedItems, nil
		},
		func(ctx context.Context, item int) error {
			if item == 1 {
				dispatchedItems <- item
				<-ctx.Done()
			}

			return nil
		},
	)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Equal(t, 1, <-dispatchedItems)
}

func TestScheduler_runOnceFetcherError(t *testing.T) {
	s := New[int]()
	expectedErr := errors.New("fetcher error")

	err := s.runOnce(
		context.Background(),
		func(ctx context.Context) ([]int, error) {
			return nil, expectedErr
		},
		func(ctx context.Context, item int) error {
			return nil
		},
	)
	require.ErrorIs(t, err, expectedErr)
}

func TestScheduler_runOnceDispatcherError(t *testing.T) {
	s := New[int]()

	expectedErr := errors.New("dispatcher error")

	err := s.runOnce(
		context.Background(),
		func(ctx context.Context) ([]int, error) {
			return []int{1, 2}, nil
		},
		func(ctx context.Context, item int) error {
			return expectedErr
		},
	)
	require.ErrorIs(t, err, expectedErr)
}

func callRunByCtx(ctx context.Context) error {
	s := New[any]()

	return s.Run(
		ctx,
		time.Second,
		func(ctx context.Context) ([]any, error) {
			return []any{}, nil
		},
		func(ctx context.Context, a any) error {
			return nil
		},
	)
}
