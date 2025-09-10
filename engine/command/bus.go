package command

import "context"

type Bus interface {
	Handle(ctx context.Context, req *Request, act ActionCallback) error
}

type Middleware func(ctx context.Context, req *Request, next ActionCallback) error

type bus struct {
	mws []Middleware
}

var passMiddleware = Middleware(func(ctx context.Context, req *Request, next ActionCallback) error {
	return next(ctx, req)
})

func NewBus(mws []Middleware) Bus {
	b := &bus{mws: mws}

	if b.mws == nil {
		b.mws = []Middleware{passMiddleware}
	}

	return b
}

func (b *bus) Handle(ctx context.Context, req *Request, act ActionCallback) error {
	return b.mws[0](ctx, req, b.getNext(1, act))
}

func (b *bus) getNext(pos int, act ActionCallback) func(context.Context, *Request) error {
	return func(ctx context.Context, req *Request) error {
		if pos >= len(b.mws) {
			return act(ctx, req)
		}

		return b.mws[pos](ctx, req, b.getNext(pos+1, act))
	}
}
