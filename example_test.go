package jsonh

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func Example_allParamsAndReturns() {
	mux := http.NewServeMux()

	type Input struct {
		I string
	}
	type Output struct {
		O string
	}

	mux.Handle("/t", ToHandler(func(w http.ResponseWriter, r *http.Request, in *Input) (*Output, error) {
		return &Output{
			O: in.I,
		}, nil
	}))

	r, err := http.NewRequest(http.MethodGet, "/t", strings.NewReader(`{"I":"from input"}`))
	if err != nil {
		panic(err)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	resp := w.Result()
	fmt.Println(resp.StatusCode)
	fmt.Println(string(w.Body.Bytes()))

	// Output:
	//200
	//{"code":200,"msg":"success","data":{"O":"from input"}}
}

func Example_hello() {
	mux := http.NewServeMux()

	mux.Handle("/hello", ToHandler(func() (string, error) {
		return "hello", nil
	}))

	r, err := http.NewRequest(http.MethodGet, "/hello", nil)
	if err != nil {
		panic(err)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	resp := w.Result()
	fmt.Println(resp.StatusCode)
	fmt.Println(string(w.Body.Bytes()))

	// Output:
	//200
	//{"code":200,"msg":"success","data":"hello"}
}

func Example_greet() {
	mux := http.NewServeMux()

	type Greet struct {
		Greet string
	}

	type GreetResponse struct {
		Reply string
	}

	mux.Handle("/greet", ToHandler(func(g *Greet) (*GreetResponse, error) {
		return &GreetResponse{
			Reply: fmt.Sprintf("Thx for %q", g.Greet),
		}, nil
	}))

	r, err := http.NewRequest(http.MethodGet, "/greet", strings.NewReader(`{"Greet":"hello"}`))
	if err != nil {
		panic(err)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	resp := w.Result()
	fmt.Println(resp.StatusCode)
	fmt.Println(string(w.Body.Bytes()))

	// Output:
	//200
	//{"code":200,"msg":"success","data":{"Reply":"Thx for \"hello\""}}
}

// if the response error implements Coder, then response the specific error code.
// Otherwise, return {"code":500,"msg":"system error"}
func Example_cerr() {
	mux := http.NewServeMux()

	mux.Handle("/cerr", ToHandler(func() error {
		return &CodeError{
			err:  errors.New("special error"),
			code: 1001,
		}
	}))

	r, err := http.NewRequest(http.MethodGet, "/cerr", nil)
	if err != nil {
		panic(err)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	resp := w.Result()
	fmt.Println(resp.StatusCode)
	fmt.Println(string(w.Body.Bytes()))

	// Output:
	//400
	//{"code":1001,"msg":"special error"}
}
