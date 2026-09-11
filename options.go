package preview

// Option adjusts how Preview evaluates a template. Options exist so a caller
// can change evaluation behavior without a new preview release, for example
// to turn off an optimization that misbehaves on a particular template.
type Option func(*options)

type options struct {
	// fullEvaluation disables resource closure pruning so every resource in
	// the root module is evaluated.
	fullEvaluation bool
}

// OptionFullEvaluation evaluates every resource in the root module instead of
// only those reachable from parameter, preset, and tag blocks. It is the escape
// hatch for the resource closure optimization: parameters, presets, and tags
// are unchanged either way, so this only matters when that optimization has a
// bug, or when Output.ModuleOutput must include outputs that read resources
// outside the closure.
func OptionFullEvaluation() Option {
	return func(o *options) {
		o.fullEvaluation = true
	}
}

func applyOptions(opts []Option) options {
	var o options
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return o
}
