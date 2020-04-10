// Author: lipixun
// Created Time : 2020-04-09 20:50:02
//
// File Name: context.go
// Description:
//

package async

import "context"

// Options defines the options used in WithAsync
type Options interface {
	Await(token *AwaitToken) Options
	Tag(tag interface{}) Options
}

// WithOptions creates a new options
func WithOptions() Options {
	return new(options)
}

// WithAsync returns a context and a token for async job
func WithAsync(ctx context.Context, opts ...Options) (context.Context, Token) {
	if len(opts) > 1 {
		panic("Too many options")
	}
	var (
		tokens []*AwaitToken
		tag    interface{}
	)

	if len(opts) > 0 {
		o := opts[0].(*options)
		tokens = append(tokens, o.awaitTokens...)
		tag = o.tag
	}
	token := getAwaitToken(ctx)
	if token != nil {
		tokens = append(tokens, token)
	}
	for _, token := range tokens {
		token.add(1)
	}
	return ctx, newToken(tokens, tag)
}

// WithAwait returns a context and wait token
func WithAwait(ctx context.Context) (context.Context, *AwaitToken) {
	return withAwaitToken(ctx)
}

type options struct {
	awaitTokens []*AwaitToken
	tag         interface{}
}

// Await adds an await token to the options
func (o *options) Await(token *AwaitToken) Options {
	if token != nil {
		for _, t := range o.awaitTokens {
			if token == t {
				return o
			}
		}
		o.awaitTokens = append(o.awaitTokens, token)
	}
	return o
}

// Tag sets the tag to the options
func (o *options) Tag(tag interface{}) Options {
	o.tag = tag
	return o
}
