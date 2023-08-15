package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"strings"
)

type HashJSONEncoder struct {
	zapcore.Encoder
}

func NewHashJSONEncoder(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
	return HashJSONEncoder{
		Encoder: zapcore.NewJSONEncoder(cfg),
	}, nil
}

func (e HashJSONEncoder) Clone() zapcore.Encoder {
	return HashJSONEncoder{
		Encoder: e.Encoder.Clone(),
	}
}

func (e HashJSONEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	fields = append(fields, zap.String("hash", calculateHash(entry.Message)))

	return e.Encoder.EncodeEntry(entry, fields)
}

func calculateHash(read string) string {
	var hashedValue uint64 = 3074457345618258791
	for _, char := range read {
		hashedValue += uint64(char)
		hashedValue *= 3074457345618258799
	}

	return strings.ToUpper(fmt.Sprintf("%x", hashedValue))
}
