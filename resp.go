package jsonh

// if the response error implements Coder, then response the specific error code.
// Otherwise, return {"code":500,"msg":"system error"}
type Coder interface {
	Code() int
}

// Resp is response
type Resp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}
