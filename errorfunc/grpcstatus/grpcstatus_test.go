package grpcstatus

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuxinx/jsonh"
	"google.golang.org/grpc/status"
)

func TestGRPCStatusErrorFunc(t *testing.T) {
	// normal error
	e1 := errors.New("error1")
	e1c := ErrorFunc(e1)
	assert.Equal(t, e1, e1c)
	_, ok := e1c.(jsonh.Coder)
	assert.Equal(t, false, ok)

	// grpc status
	e2 := status.Error(111, "error2")
	_, ok = e2.(jsonh.Coder)
	assert.Equal(t, false, ok)
	e2c := ErrorFunc(e2)
	e2cjc, ok := e2c.(jsonh.Coder)
	assert.Equal(t, true, ok)
	assert.Equal(t, 111, e2cjc.Code())
	assert.Equal(t, "error2", e2c.Error())
}
