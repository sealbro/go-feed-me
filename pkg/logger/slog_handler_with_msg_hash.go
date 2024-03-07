package logger

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

const HashKey = "hash"

type MsgHashHandler struct {
	Next slog.Handler
}

func (h *MsgHashHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(slog.String(HashKey, calculateHash(r.Message)))

	return h.Next.Handle(ctx, r)
}

func (h *MsgHashHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Next.Enabled(ctx, level)
}

func (h *MsgHashHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MsgHashHandler{Next: h.Next.WithAttrs(attrs)}
}

func (h *MsgHashHandler) WithGroup(name string) slog.Handler {
	return &MsgHashHandler{Next: h.Next.WithGroup(name)}
}

func calculateHash(read string) string {
	var hashedValue uint64 = 3074457345618258791
	for _, char := range read {
		hashedValue += uint64(char)
		hashedValue *= 3074457345618258799
	}

	return strings.ToUpper(fmt.Sprintf("%x", hashedValue))
}
