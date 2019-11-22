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

	return cerr.New(int(st.Code()), strings.Split(e.Error(), "desc = ")[1])
}
