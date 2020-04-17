// Author: lipixun
// Created Time : 2020-04-09 20:51:04
//
// File Name: core.go
// Description:
//

package async

import (
	"context"
	"strings"
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
	// Get the conncurent number of this async job
	//	0 will be returned if not in an concurent context
	ConcurrentNum() int
}

//
//
// AwaitToken
//
//

// AwaitToken is used to manage difference async job synchronous requirements
type AwaitToken struct {
	m            sync.Mutex
	wg           sync.WaitGroup
	states       []State
	doneCallback func(*token)
}

// Wait for the jobs which is connected to this token completed
func (t *AwaitToken) Wait() []State {
	t.wg.Wait()
	return t.states
}

// WaitError waits for the jobs which is connected to this token completed and returns error
func (t *AwaitToken) WaitError() error {
	t.wg.Wait()
	var errs []error
	for _, state := range t.states {
		if state.Error() != nil {
			errs = append(errs, state.Error())
		}
	}
	if len(errs) > 0 {
		return newCompoundError(errs...)
	}
	return nil
}

func (t *AwaitToken) add(n int) {
	t.wg.Add(n)
}

func (t *AwaitToken) done(tk *token) { //err error, tag interface{}) {
	t.m.Lock()
	t.states = append(t.states, &state{tk.err, tk.tag})
	t.m.Unlock()
	t.wg.Done()
	if t.doneCallback != nil {
		t.doneCallback(tk)
	}
}

type awaitTokenContextKey struct{}

func getAwaitToken(ctx context.Context) *AwaitToken {
	token, ok := ctx.Value(awaitTokenContextKey{}).(*AwaitToken)
	if !ok {
		return nil
	}
	return token
}

func withAwaitToken(ctx context.Context, doneCallback func(*token)) (context.Context, *AwaitToken) {
	token := AwaitToken{
		doneCallback: doneCallback,
	}
	return context.WithValue(ctx, awaitTokenContextKey{}, &token), &token
}

//
//
// Token
//
//

type token struct {
	awaitTokens  []*AwaitToken
	done         bool
	err          error
	tag          interface{}
	cno          int
	doneCallback func(*token)
}

func newToken(awaitTokens []*AwaitToken, tag interface{}, cno int, doneCallback func(*token)) *token {
	return &token{awaitTokens: awaitTokens, tag: tag, cno: cno, doneCallback: doneCallback}
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
		token.done(t)
	}
	t.done = true
	if t.doneCallback != nil {
		t.doneCallback(t)
	}
}

func (t *token) SetError(err error) {
	if err != nil {
		t.err = err
	}
}

// Get the conncurent number of this async job
func (t *token) ConcurrentNum() int {
	return t.cno
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

//
//
// Error
//
//

// CompoundError defines the compound error
type CompoundError struct {
	s    string
	errs []error
}

func newCompoundError(err ...error) *CompoundError {
	var errStrs []string
	for _, e := range err {
		errStrs = append(errStrs, e.Error())
	}
	return &CompoundError{strings.Join(errStrs, "\n"), err}
}

// Errors returns the errors
func (e *CompoundError) Errors() []error {
	return e.errs
}

// Errors returns the error
func (e *CompoundError) Error() string {
	return e.s
}
