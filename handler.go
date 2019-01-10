// Package jsonh convert Go func to http.Handler that handle json request and response json
package jsonh

import (
	"net/http"
	"reflect"

	"github.com/json-iterator/go"
	"github.com/xuxinx/cerr"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()

var systemErrorResp, _ = jsoniter.Marshal(&Resp{
	Code: http.StatusInternalServerError,
	Msg:  "system error",
})
var requestParamErrorResp, _ = jsoniter.Marshal(&Resp{
	Code: http.StatusUnprocessableEntity,
	Msg:  "input error",
})
var noDataSuccessResp, _ = jsoniter.Marshal(&Resp{
	Code: http.StatusOK,
	Msg:  "success",
})

// ToHandler convert f to http.Handler.
//
// f should be like `func(r, input) (output, error)`,
// input and output can be omitted.
// r is *http.Request.
// input can be struct or struct pointer.
// output can be any type
func ToHandler(f interface{}) http.Handler {
	fV := reflect.ValueOf(f)
	fT := reflect.TypeOf(f)

	var inT reflect.Type
	var isInPtr bool
	numIn := fT.NumIn()
	if numIn == 0 {
		panic("missing *http.Request")
	}
	if numIn > 2 {
		panic("too many params")
	}
	if fT.In(0) != reflect.TypeOf((*http.Request)(nil)) {
		panic("fist param is not *http.Request")
	}
	if numIn == 2 {
		inT = fT.In(1)
		if inT.Kind() != reflect.Struct &&
			!(inT.Kind() == reflect.Ptr && inT.Elem().Kind() == reflect.Struct) {
			panic("input is not struct or struct pointer")
		}

		if inT.Kind() == reflect.Ptr {
			isInPtr = true
		}
	}

	numOut := fT.NumOut()
	if numOut == 0 {
		panic("missing return error")
	}
	if numIn > 2 {
		panic("too many return params")
	}
	if !fT.Out(numOut - 1).Implements(errorType) {
		panic("last return param is not error")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var err error

		var callResp []reflect.Value
		{
			callIn := make([]reflect.Value, 1, 2)
			callIn[0] = reflect.ValueOf(r)
			if inT != nil {
				var in interface{}
				if isInPtr {
					in = reflect.New(inT.Elem()).Interface()
				} else {
					in = reflect.New(inT).Interface()
				}

				err = jsoniter.NewDecoder(r.Body).Decode(in)
				if err != nil {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write(requestParamErrorResp)
					return
				}

				var inV reflect.Value
				if isInPtr {
					inV = reflect.ValueOf(in)
				} else {
					inV = reflect.ValueOf(in).Elem()
				}

				callIn = append(callIn, inV)
			}
			callResp = fV.Call(callIn)
		}

		respErr := callResp[len(callResp)-1]
		if !respErr.IsNil() {
			cErr, ok := respErr.Interface().(cerr.CodeError)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(systemErrorResp)
				return
			}

			resBuf, _ := jsoniter.Marshal(&Resp{
				Code: cErr.Code(),
				Msg:  cErr.Error(),
			})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(resBuf)
			return
		}

		if len(callResp) > 1 {
			resBuf, _ := jsoniter.Marshal(&Resp{
				Code: http.StatusOK,
				Msg:  "success",
				Data: callResp[0].Interface(),
			})
			w.WriteHeader(http.StatusOK)
			w.Write(resBuf)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(noDataSuccessResp)
	})
}
