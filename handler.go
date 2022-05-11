// Package jsonh convert Go func to http.Handler that handle json request and response json
package jsonh

import (
	"encoding/json"
	"net/http"
	"reflect"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()
var httpResponseWriterType = reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
var httpRequestType = reflect.TypeOf((*http.Request)(nil))

var systemErrorResp, _ = json.Marshal(&Resp{
	Code: http.StatusInternalServerError,
	Msg:  "system error",
})
var unprocessableRequestParamsResp, _ = json.Marshal(&Resp{
	Code: http.StatusUnprocessableEntity,
	Msg:  "unprocessable request params",
})
var successWithoutDataResp, _ = json.Marshal(&Resp{
	Code: http.StatusOK,
	Msg:  "success",
})

// ToHandler convert f to http.Handler.
//
// f should be like `func(w, r, input) (output, error)`,
// w, r, input and output can be omitted.
// w is http.ResponseWriter.
// r is *http.Request.
// input is struct pointer.
// output can be any type.
func ToHandler(f interface{}) http.Handler {
	return ToHandlerWithErrorFunc(f, nil)
}

func ToHandlerFunc(f interface{}) http.HandlerFunc {
	return ToHandlerFuncWithErrorFunc(f, nil)
}

// ErrorFunc is used to make error impl type Coder
// e.g. gRPC status error
// import (
// 	"github.com/xuxinx/jsonh/cerr"
// 	"google.golang.org/grpc/status"
// )
//
// func ErrorFunc(e error) error {
// 	st, ok := status.FromError(e)
// 	if !ok {
// 		return e
// 	}
//
// 	return cerr.New(int(st.Code()), e.Error())
// }
type ErrorFunc func(error) error

func ToHandlerWithErrorFunc(f interface{}, ef ErrorFunc) http.Handler {
	return ToHandlerFuncWithErrorFunc(f, ef)
}

func ToHandlerFuncWithErrorFunc(f interface{}, ef ErrorFunc) http.HandlerFunc {
	if f == nil {
		panic("f cannot be nil")
	}

	fV := reflect.ValueOf(f)
	fT := fV.Type()
	if fT.Kind() != reflect.Func {
		panic("f is not a function")
	}

	var hasW bool
	var hasR bool
	var inT reflect.Type
	numIn := fT.NumIn()
	if numIn > 3 {
		panic("too many params")
	}
	for i := 0; i < numIn; i++ {
		pT := fT.In(i)
		if pT == httpResponseWriterType {
			if hasW {
				panic("too many param w")
			}
			if hasR || inT != nil {
				panic("param w wrong place")
			}
			hasW = true
		} else if pT == httpRequestType {
			if hasR {
				panic("too many param r")
			}
			if inT != nil {
				panic("param r wrong place")
			}
			hasR = true
		} else {
			if inT != nil {
				panic("too many param input")
			}
			inT = pT
			if inT.Kind() != reflect.Ptr || inT.Elem().Kind() != reflect.Struct {
				panic("param input is not struct pointer")
			}
		}
	}

	numOut := fT.NumOut()
	if numOut == 0 {
		panic("missing return error")
	}
	if numOut > 2 {
		panic("too many return params")
	}
	if !fT.Out(numOut - 1).Implements(errorType) {
		panic("last return param is not error")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		w.Header().Set("Content-Type", "application/json")

		var callResp []reflect.Value
		{
			callIn := make([]reflect.Value, 0, 3)
			if hasW {
				callIn = append(callIn, reflect.ValueOf(w))
			}
			if hasR {
				callIn = append(callIn, reflect.ValueOf(r))
			}
			if inT != nil {
				in := reflect.New(inT.Elem()).Interface()
				err = json.NewDecoder(r.Body).Decode(in)
				if err != nil {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write(unprocessableRequestParamsResp)
					return
				}

				callIn = append(callIn, reflect.ValueOf(in))
			}
			callResp = fV.Call(callIn)
		}

		respErr := callResp[len(callResp)-1]
		if !respErr.IsNil() {
			rErr := respErr.Interface().(error)
			if ef != nil {
				rErr = ef(rErr)
			}
			cErr, ok := rErr.(Coder)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(systemErrorResp)
				return
			}

			resBuf, _ := json.Marshal(&Resp{
				Code: cErr.Code(),
				Msg:  rErr.Error(),
			})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(resBuf)
			return
		}

		if len(callResp) > 1 {
			resBuf, _ := json.Marshal(&Resp{
				Code: http.StatusOK,
				Msg:  "success",
				Data: callResp[0].Interface(),
			})
			w.WriteHeader(http.StatusOK)
			w.Write(resBuf)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(successWithoutDataResp)
	}
}
