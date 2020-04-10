// Author: lipixun
// Created Time : 2020-04-09 20:51:04
//
// File Name: core.go
// Description:
//

package async

import (
	"context"
	"sync"
)

// State defines the interface of the state of a async job
type State interface {
	Error() error
	Tag() interface{}
}

// Token defines the interface which should be used to set async job state
type Token interface {
	Done(error)
}

// AWaitToken is used to manage difference async job synchronous requirements
type AWaitToken struct {
	m      sync.Mutex
	wg     sync.WaitGroup
	states []State
}

// Wait for the jobs which is connected to this token completed
func (t *AWaitToken) Wait() []State {
	t.wg.Wait()
	return t.states
}

func (t *AWaitToken) add(n int) {
	t.wg.Add(n)
}

func (t *AWaitToken) done(err error, tag interface{}) {
	t.m.Lock()
	t.states = append(t.states, &state{err, tag})
	t.m.Unlock()
	t.wg.Done()
}

type awaitTokenContextKey struct{}

func getAWaitToken(ctx context.Context) *AWaitToken {
	token, ok := ctx.Value(awaitTokenContextKey{}).(*AWaitToken)
	if !ok {
		return nil
	}
	return token
}

func withAWaitToken(ctx context.Context) (context.Context, *AWaitToken) {
	var token AWaitToken
	return context.WithValue(ctx, awaitTokenContextKey{}, &token), &token
}

type token struct {
	awaitTokens []*AWaitToken
	tag         interface{}
}

func newToken(awaitTokens []*AWaitToken, tag interface{}) *token {
	return &token{awaitTokens, tag}
}

func (t *token) Done(err error) {
	for _, token := range t.awaitTokens {
		token.done(err, t.tag)
	}
}

type state struct {
	err error
	tag interface{}
}

func (s *state) Error() error {
	return s.err
}

func (s *state) Tag() interface{} {
	return s.tag
}
