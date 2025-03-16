package logger

type Field struct {
	Key string
	Val any
}

type Logger interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
}

func Error(err error) Field {
	return Field{
		Key: "error",
		Val: err,
	}
}

func Int(key string, i int) Field {
	return Field{
		Key: key,
		Val: i,
	}
}

func Int32(key string, val int32) Field {
	return Field{Key: key, Val: val}
}

func Int64(key string, i int64) Field {
	return Field{
		Key: key,
		Val: i,
	}
}

func String(key string, str string) Field {
	return Field{
		Key: key,
		Val: str,
	}
}
