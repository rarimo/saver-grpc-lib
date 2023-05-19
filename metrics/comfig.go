package metrics

import (
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Profilerer interface {
	Profiler() Profiler
}

type profilerer struct {
	getter kv.Getter
	once   comfig.Once
}

func New(getter kv.Getter) Profilerer {
	return &profilerer{
		getter: getter,
	}
}

func (p *profilerer) Profiler() Profiler {
	return p.once.Do(func() interface{} {
		profiler := Profiler{}

		if err := figure.Out(&profiler).From(kv.MustGetStringMap(p.getter, "profiler")).Please(); err != nil {
			panic(errors.Wrap(err, "failed to figure out profiler config"))
		}

		return profiler
	}).(Profiler)
}
