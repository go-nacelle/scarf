package logging

import (
	"context"

	"github.com/efritz/nacelle"
)

type (
	Decorator struct {
		extractors []LogFieldExtractor
	}

	LogFieldExtractor func(ctx context.Context, fields nacelle.LogFields)
)

func NewDecorator(extractors ...LogFieldExtractor) *Decorator {
	return &Decorator{
		extractors: extractors,
	}
}

func (d *Decorator) Register(extractor LogFieldExtractor) {
	d.extractors = append(d.extractors, extractor)
}

func (d *Decorator) Decorate(ctx context.Context, logger nacelle.Logger) nacelle.Logger {
	fields := nacelle.LogFields{}
	for _, extractor := range d.extractors {
		extractor(ctx, fields)
	}

	return logger.WithFields(fields)
}
