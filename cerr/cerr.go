package cerr

type CodeError interface {
	Error() string
	Code() int

	private()
}

func New(code int, msg string) CodeError {
	return &cerr{
		code: code,
		msg:  msg,
	}
}

type cerr struct {
	code int
	msg  string
}

func (e *cerr) Error() string {
	return e.msg
}

func (e *cerr) Code() int {
	return e.code
}

func (*cerr) private() {}
