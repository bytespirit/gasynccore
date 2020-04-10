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
	// Done will record the first non-nil error
	// 	NOTE:
	//		This method MUST ONLY be called once
	Done(...error)
	// Set error. This method is allowed to call multiple times. And only the
	//	last non-nil error will be saved
	// NOTE: This method will not mark the job is done.
	SetError(error)
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
	done        bool
	err         error
	tag         interface{}
}

func newToken(awaitTokens []*AWaitToken, tag interface{}) *token {
	return &token{awaitTokens: awaitTokens, tag: tag}
}

func (t *token) Done(errs ...error) {
	if t.done {
		panic("Done is called too many times.")
	}
	for _, e := range errs {
		if e != nil {
			t.err = e
		}
	}
	for _, token := range t.awaitTokens {
		token.done(t.err, t.tag)
	}
	t.done = true
}

func (t *token) SetError(err error) {
	if err != nil {
		t.err = err
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
