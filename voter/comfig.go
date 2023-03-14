package voter

import (
	"time"

	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type SubscriberConfig struct {
	MinRetryPeriod time.Duration `fig:"min_retry_period"`
	MaxRetryPeriod time.Duration `fig:"max_retry_period"`
}

type Subscriberer interface {
	Subscriber() SubscriberConfig
}

func NewSubscriberer(getter kv.Getter) Subscriberer {
	return &subscriberer{
		getter: getter,
	}
}

type subscriberer struct {
	getter kv.Getter
	once   comfig.Once
}

func (s *subscriberer) Subscriber() SubscriberConfig {
	return s.once.Do(func() interface{} {
		cfg := SubscriberConfig{
			MinRetryPeriod: 1 * time.Second,
			MaxRetryPeriod: 10 * time.Second,
		}

		err := figure.
			Out(&cfg).
			From(kv.MustGetStringMap(s.getter, "subscriber")).
			Please()

		if err != nil {
			panic(errors.Wrap(err, "failed to figure out subscriber config"))
		}

		return cfg
	}).(SubscriberConfig)
}
