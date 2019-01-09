package jsonh

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/xuxinx/cerr"
)

type Obj struct {
	F string `json:"f"`
}

type Input struct {
	I1 string `json:"i1"`
	I2 int    `json:"i2"`
	I3 *Obj   `json:"i3"`
}

type Output struct {
	O1 string `json:"o1"`
	O2 int    `json:"o2"`
	O3 *Obj   `json:"o3"`
}

func TestToHandler(t *testing.T) {
	cases := []struct {
		f     interface{}
		input interface{}

		expCode int
		expResp string
	}{
		// success
		{
			f: func(r *http.Request, input *Input) (*Output, error) {
				if input.I1 == "b" {
					return nil, cerr.New(1000, "err b")
				}
				if input.I1 == "c" {
					return nil, errors.New("err c")
				}

				return &Output{
					O1: input.I1,
					O2: input.I2,
					O3: input.I3,
				}, nil
			},
			input: &Input{
				I1: "a",
				I2: 11,
				I3: &Obj{
					F: "ff",
				},
			},
			expCode: http.StatusOK,
			expResp: `{"code":200,"msg":"success","data":{"o1":"a","o2":11,"o3":{"f":"ff"}}}`,
		},
		// struct input
		{
			f: func(r *http.Request, input Input) (*Output, error) {
				if input.I1 == "b" {
					return nil, cerr.New(1000, "err b")
				}
				if input.I1 == "c" {
					return nil, errors.New("err c")
				}

				return &Output{
					O1: input.I1,
					O2: input.I2,
					O3: input.I3,
				}, nil
			},
			input: Input{
				I1: "a",
				I2: 11,
				I3: &Obj{
					F: "ff",
				},
			},
			expCode: http.StatusOK,
			expResp: `{"code":200,"msg":"success","data":{"o1":"a","o2":11,"o3":{"f":"ff"}}}`,
		},
		// custom error
		{
			f: func(r *http.Request, input *Input) (*Output, error) {
				if input.I1 == "b" {
					return nil, cerr.New(1000, "err b")
				}
				if input.I1 == "c" {
					return nil, errors.New("err c")
				}

				return &Output{
					O1: input.I1,
					O2: input.I2,
					O3: input.I3,
				}, nil
			},
			input: &Input{
				I1: "b",
				I2: 11,
				I3: &Obj{
					F: "ff",
				},
			},
			expCode: http.StatusBadRequest,
			expResp: `{"code":1000,"msg":"err b"}`,
		},
		// unknown error
		{
			f: func(r *http.Request, input *Input) (*Output, error) {
				if input.I1 == "b" {
					return nil, cerr.New(1000, "err b")
				}
				if input.I1 == "c" {
					return nil, errors.New("err c")
				}

				return &Output{
					O1: input.I1,
					O2: input.I2,
					O3: input.I3,
				}, nil
			},
			input: &Input{
				I1: "c",
				I2: 11,
				I3: &Obj{
					F: "ff",
				},
			},
			expCode: http.StatusInternalServerError,
			expResp: string(systemErrorResp),
		},

		// no input
		{
			f: func(r *http.Request) (*Output, error) {
				return &Output{
					O1: "a",
					O2: 11,
					O3: &Obj{
						F: "ff",
					},
				}, nil
			},
			input:   nil,
			expCode: http.StatusOK,
			expResp: `{"code":200,"msg":"success","data":{"o1":"a","o2":11,"o3":{"f":"ff"}}}`,
		},

		// no output
		{
			f: func(r *http.Request, input *Input) error {
				return nil
			},
			input: &Input{
				I1: "a",
			},
			expCode: http.StatusOK,
			expResp: `{"code":200,"msg":"success"}`,
		},
	}

	var err error
	for _, c := range cases {
		mux := http.NewServeMux()
		mux.Handle("/t", ToHandler(c.f))

		var r *http.Request
		if c.input == nil {
			r, err = http.NewRequest(http.MethodGet, "/t", nil)
		} else {
			input, _ := jsoniter.MarshalToString(c.input)
			r, err = http.NewRequest(http.MethodGet, "/t", strings.NewReader(input))
		}
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		resp := w.Result()
		assert.Equal(t, c.expCode, resp.StatusCode)
		assert.Equal(t, c.expResp, string(w.Body.Bytes()))
	}
}

func TestToHandlerFunc_UnprocessableEntity(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/t", ToHandler(func(r *http.Request, input *Input) error {
		return nil
	}))

	r, err := http.NewRequest(http.MethodGet, "/t", strings.NewReader("a"))
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.Equal(t, string(requestParamErrorResp), string(w.Body.Bytes()))
}

var (
	f = func(r *http.Request, input *Input) (output *Output, err error) {
		return &Output{
			O1: input.I1,
			O2: input.I2,
			O3: input.I3,
		}, nil
	}
	inputStr = `{"i1":"a","i2":1,"i3":{"f":"ff"}}`
)

func BenchmarkToHandler(b *testing.B) {
	var err error
	mux := http.NewServeMux()
	mux.Handle("/t", ToHandler(f))

	var r *http.Request
	var w *httptest.ResponseRecorder
	for i := 0; i < b.N; i++ {
		inReader := strings.NewReader(inputStr)
		r, err = http.NewRequest(http.MethodGet, "/t", inReader)
		if err != nil {
			panic(err)
		}
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		//fmt.Println("1", w.Body.String())
	}
}

func BenchmarkStdHTTPHandler(b *testing.B) {
	var err error
	mux := http.NewServeMux()
	mux.Handle("/t", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyData, _ := ioutil.ReadAll(r.Body)
		in := &Input{}
		jsoniter.Unmarshal(bodyData, in)
		o, _ := f(r, in)
		w.WriteHeader(http.StatusOK)
		jsoniter.NewEncoder(w).Encode(o)
	}))

	var r *http.Request
	var w *httptest.ResponseRecorder
	for i := 0; i < b.N; i++ {
		inReader := strings.NewReader(inputStr)
		r, err = http.NewRequest(http.MethodGet, "/t", inReader)
		if err != nil {
			panic(err)
		}
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		//fmt.Println("2", w.Body.String())
	}
}

//$ go test -bench=. -run=404
//goos: darwin
//goarch: amd64
//pkg: github.com/xuxinx/goutils/jsonh
//BenchmarkToHandler-8        	  300000	      4035 ns/op
//BenchmarkStdHTTPHandler-8   	  500000	      3072 ns/op
