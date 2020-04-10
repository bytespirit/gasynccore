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
	AWait(token *AWaitToken) Options
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
		tokens []*AWaitToken
		tag    interface{}
	)

	if len(opts) > 0 {
		o := opts[0].(*options)
		tokens = append(tokens, o.awaitTokens...)
		tag = o.tag
	}
	token := getAWaitToken(ctx)
	if token != nil {
		tokens = append(tokens, token)
	}
	for _, token := range tokens {
		token.add(1)
	}
	return ctx, newToken(tokens, tag)
}

// WithAWait returns a context and wait token
func WithAWait(ctx context.Context) (context.Context, *AWaitToken) {
	return withAWaitToken(ctx)
}

type options struct {
	awaitTokens []*AWaitToken
	tag         interface{}
}

// AWait adds an await token to the options
func (o *options) AWait(token *AWaitToken) Options {
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
