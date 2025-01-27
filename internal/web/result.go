package web

type Result struct {
	Code uint16 `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
