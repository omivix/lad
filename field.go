package lad

import (
	"go.uber.org/zap/zapcore"
)

func Binary(key string, val []byte) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.BinaryType, Interface: val}
}

func String(key string, val string) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.StringType, String: val}
}

func Reflect(key string, val interface{}) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.ReflectType, Interface: val}
}
