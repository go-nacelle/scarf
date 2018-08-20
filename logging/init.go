package logging

import "github.com/efritz/nacelle"

type Initializer struct {
	Services   nacelle.ServiceContainer `service:"container"`
	extractors []LogFieldExtractor
}

const ServiceName = "logger-decorator"

func NewInitializer(extractors []LogFieldExtractor) *Initializer {
	return &Initializer{
		extractors: extractors,
	}
}

func (i *Initializer) Init(config nacelle.Config) error {
	return i.Services.Set(ServiceName, NewDecorator(i.extractors...))
}
