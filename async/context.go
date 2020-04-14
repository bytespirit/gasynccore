// Author: lipixun
// Created Time : 2020-04-09 20:50:02
//
// File Name: context.go
// Description:
//

package async

import "context"

// WithAsync returns a context and a token for async job
func WithAsync(ctx context.Context, opts ...Option) (context.Context, Token) {
	var (
		o      options
		tokens []*AwaitToken
	)

	for _, opt := range opts {
		opt.apply(&o)
	}

	tokens = append(tokens, o.awaitTokens...)
	token := getAwaitToken(ctx)
	if token != nil {
		tokens = append(tokens, token)
	}
	for _, token := range tokens {
		token.add(1)
	}
	return ctx, newToken(tokens, o.tag, o.cno, o.doneCallback)
}

// WithAwait returns a context and wait token
func WithAwait(ctx context.Context, opts ...Option) (context.Context, *AwaitToken) {
	var o options
	for _, opt := range opts {
		opt.apply(&o)
	}
	return withAwaitToken(ctx, o.doneCallback)
}

// Option defines the option used in WithAsync
type Option interface {
	apply(*options)
}

type options struct {
	awaitTokens  []*AwaitToken
	tag          interface{}
	cno          int // Concurrent number
	doneCallback func(*token)
}

type funcOption struct {
	f func(*options)
}

func newFuncOption(f func(*options)) funcOption {
	return funcOption{f}
}

func (o funcOption) apply(opts *options) {
	o.f(opts)
}

// Await adds an await token to the options
func Await(token *AwaitToken) Option {
	return newFuncOption(func(o *options) {
		if token != nil {
			o.awaitTokens = append(o.awaitTokens, token)
		}
	})
}

// Tag sets the tag to the options
func Tag(tag interface{}) Option {
	return newFuncOption(func(o *options) {
		o.tag = tag
	})
}

func cno(cno int) Option {
	return newFuncOption(func(o *options) {
		o.cno = cno
	})
}

func doneCallback(cb func(*token)) Option {
	return newFuncOption(func(o *options) {
		o.doneCallback = cb
	})
}
