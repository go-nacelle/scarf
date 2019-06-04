package scarf

import (
	"context"

	"github.com/go-nacelle/nacelle"

	"github.com/go-nacelle/scarf/logging"
)

var DefaultExtractors = []logging.LogFieldExtractor{
	PopulateClientIDField,
	PopulateRemoteAddrField,
}

func PopulateClientIDField(ctx context.Context, fields nacelle.LogFields) {
	if clientID := GetClientID(ctx); clientID != "" {
		fields["client_id"] = clientID
	}
}

func PopulateRemoteAddrField(ctx context.Context, fields nacelle.LogFields) {
	if remoteAddr := GetRemoteAddr(ctx); remoteAddr != "" {
		fields["remote_addr"] = remoteAddr
	}
}
