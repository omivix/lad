package lad

import (
	"math"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Any(key string, value interface{}) zapcore.Field {
	return zap.Any(key, value)
}

func Binary(key string, val []byte) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.BinaryType, Interface: val}
}

func Bool(key string, val bool) zapcore.Field {
	var ival int64
	if val {
		ival = 1
	}
	return zapcore.Field{Key: key, Type: zapcore.BoolType, Integer: ival}
}

func Dict(key string, val ...zapcore.Field) zapcore.Field {
	return zap.Dict(key, val...)
}

func Error(err error) zapcore.Field {
	return zap.Error(err)
}

func Float32(key string, val float32) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Float32Type, Integer: int64(math.Float32bits(val))}
}

func Float64(key string, val float64) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Float64Type, Integer: int64(math.Float64bits(val))}
}

func Int(key string, val int) zapcore.Field {
	return Int64(key, int64(val))
}

func Int8(key string, val int8) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Int8Type, Integer: int64(val)}
}

func Int16(key string, val int8) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Int16Type, Integer: int64(val)}
}

func Int32(key string, val int64) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Int32Type, Integer: int64(val)}
}

func Int64(key string, val int64) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Int64Type, Integer: val}
}

func Reflect(key string, val interface{}) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.ReflectType, Interface: val}
}

func String(key string, val string) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.StringType, String: val}
}
