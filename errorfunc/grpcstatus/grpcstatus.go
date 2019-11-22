package grpcstatus

import (
	"strings"

	"github.com/xuxinx/cerr"
	"google.golang.org/grpc/status"
)

func ErrorFunc(e error) error {
	st, ok := status.FromError(e)
	if !ok {
		return e
	}

	// grpc status error format
	// rpc error: code = Code(__code__) desc = __desc__
	msg := e.Error()
	{
		msgs := strings.Split(msg, " desc = ")
		if len(msgs) > 1 {
			msg = msgs[1]
		}
	}

	return cerr.New(int(st.Code()), msg)
}
