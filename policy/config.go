package policy

type Option func(*Policy)

func WithAdapter(a any) Option {
	return func(p *Policy) {
		p.adapter = a
	}
}
