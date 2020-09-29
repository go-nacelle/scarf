package scarf

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/stats"
)

type StatsHandler struct {
	handler stats.Handler
}

type tokenClientID string

type tokenRemoteAddr string

var (
	TokenClientID   = tokenClientID("scarf.client_id")
	TokenRemoteAddr = tokenRemoteAddr("scarf.remote_addr")
)

func GetClientID(ctx context.Context) string {
	if val, ok := ctx.Value(TokenClientID).(string); ok {
		return val
	}

	return ""
}

func GetRemoteAddr(ctx context.Context) string {
	if val, ok := ctx.Value(TokenRemoteAddr).(string); ok {
		return val
	}

	return ""
}

func NewStatsHandler(handler stats.Handler) stats.Handler {
	return &StatsHandler{
		handler: handler,
	}
}

func (h *StatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	raw, err := uuid.NewRandom()
	if err != nil {
		// TODO - what to do with this?
	}

	clientID := raw.String()
	ctx = context.WithValue(ctx, TokenClientID, clientID)
	ctx = context.WithValue(ctx, TokenRemoteAddr, info.RemoteAddr)

	if h.handler == nil {
		return ctx
	}

	return h.handler.TagConn(ctx, info)
}

func (h *StatsHandler) HandleConn(ctx context.Context, stats stats.ConnStats) {
	if h.handler != nil {
		h.handler.HandleConn(ctx, stats)
	}
}

func (h *StatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	if h.handler == nil {
		return ctx

	}

	return h.handler.TagRPC(ctx, info)
}

func (h *StatsHandler) HandleRPC(ctx context.Context, stats stats.RPCStats) {
	if h.handler != nil {
		h.handler.HandleRPC(ctx, stats)
	}
}
