package logger

type NopLogger struct {
}

func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

func (n *NopLogger) Debug(_ string, _ ...Field) {

}

func (n *NopLogger) Info(_ string, _ ...Field) {

}

func (n *NopLogger) Warn(_ string, _ ...Field) {

}

func (n *NopLogger) Error(_ string, _ ...Field) {

}
