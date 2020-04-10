// Author: lipixun
// Created Time : 2020-04-10 01:32:03
//
// File Name: test.go
// Description:
//

package async

import (
	"context"
	"errors"
	"testing"
	"time"
)

func Test_BasicUsage(t *testing.T) {
	ctx, awaitToken := WithAWait(context.Background())

	value := 0

	go func(ctx context.Context, token Token) {
		defer func() {
			token.Done(nil)
		}()
		time.Sleep(time.Second)
		value = 1
	}(WithAsync(ctx))

	states := awaitToken.Wait()
	if len(states) != 1 {
		t.Fatalf("States must be 1. Actually: %v", len(states))
	}
	if states[0].Error() != nil {
		t.Fatalf("States error must be nil. Actually: %v", states[0].Error())
	}
	if value != 1 {
		t.Fatalf("Value must be 1. Actually: %v", value)
	}
}

func Test_WithoutAsync(t *testing.T) {
	ctx, awaitToken := WithAWait(context.Background())

	value := 0

	go func(ctx context.Context) {
		time.Sleep(time.Second)
		value = 1
	}(ctx)

	states := awaitToken.Wait()
	if len(states) != 0 {
		t.Fatalf("States must be zero. Actually: %v", len(states))
	}
	if value != 0 {
		t.Fatalf("Value must be 0. Actually: %v", value)
	}
}

func Test_WithTagErrorAndToken(t *testing.T) {
	ctx, _ := WithAWait(context.Background())

	var (
		value       int
		awaitToken2 AWaitToken
	)

	go func(ctx context.Context, token Token) {
		defer func() {
			token.Done(errors.New("123"))
		}()
		time.Sleep(time.Second)
		value = 1
	}(WithAsync(ctx, WithOptions().AWait(&awaitToken2).Tag(666)))

	states := awaitToken2.Wait()
	if len(states) != 1 {
		t.Fatalf("States must be 1. Actually: %v", len(states))
	}
	if states[0].Error() == nil || states[0].Error().Error() != "123" {
		t.Fatalf("States error must be 123. Actually: %v", states[0].Error())
	}
	if states[0].Tag() != 666 {
		t.Fatalf("States tag must be 666. Actually: %v", states[0].Tag())
	}
	if value != 1 {
		t.Fatalf("Value must be 1. Actually: %v", value)
	}
}
