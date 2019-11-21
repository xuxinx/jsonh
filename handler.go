// Package jsonh convert Go func to http.Handler that handle json request and response json
package jsonh

import (
	"encoding/json"
	"net/http"
	"reflect"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()
var httpResponseWriteType = reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
var httpRequestType = reflect.TypeOf((*http.Request)(nil))

var systemErrorResp, _ = json.Marshal(&Resp{
	Code: http.StatusInternalServerError,
	Msg:  "system error",
})
var requestParamErrorResp, _ = json.Marshal(&Resp{
	Code: http.StatusUnprocessableEntity,
	Msg:  "input error",
})
var noDataSuccessResp, _ = json.Marshal(&Resp{
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

// Can make error to impl type Coder by ErrorFunc
type ErrorFunc func(error) error

func ToHandlerWithErrorFunc(f interface{}, ef ErrorFunc) http.Handler {
	fV := reflect.ValueOf(f)
	fT := reflect.TypeOf(f)

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
		if fT.In(i) == httpResponseWriteType {
			if hasW {
				panic("too many param w")
			}
			if hasR || inT != nil {
				panic("param w wrong place")
			}
			hasW = true
		} else if fT.In(i) == httpRequestType {
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
			inT = fT.In(i)
			if !(inT.Kind() == reflect.Ptr && inT.Elem().Kind() == reflect.Struct) {
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

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var err error

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
					w.Write(requestParamErrorResp)
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
				Msg:  respErr.Interface().(error).Error(),
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
		w.Write(noDataSuccessResp)
	})
}
