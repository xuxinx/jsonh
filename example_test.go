package jsonh_test

import (
	"fmt"
	"github.com/prometheus/common/log"
	"github.com/xuxinx/cerr"
	"github.com/xuxinx/jsonh"
	"net/http"
)

func Example_hello() {
	http.Handle("/hello", jsonh.ToHandler(func(_ *http.Request) (string, error) {
		return "hello", nil
	}))

	log.Fatal(http.ListenAndServe(":8080", nil))

	// Output:
	//$ curl -i 'localhost:8080/hello'
	//HTTP/1.1 200 OK
	//Content-Type: application/json
	//Date: Thu, 10 Jan 2019 05:03:35 GMT
	//Content-Length: 43
	//
	//{"code":200,"msg":"success","data":"hello"}%
}

func Example_greet() {
	type Greet struct {
		Greet string
	}

	type GreetResponse struct {
		Reply string
	}

	http.Handle("/greet", jsonh.ToHandler(func(_ *http.Request, g *Greet) (*GreetResponse, error) {
		return &GreetResponse{
			Reply: fmt.Sprintf("Thx for %q", g.Greet),
		}, nil
	}))

	log.Fatal(http.ListenAndServe(":8080", nil))

	// Output:
	//$ curl -i -d '{"Greet":"hello"}' 'localhost:8080/greet'
	//HTTP/1.1 200 OK
	//Content-Type: application/json
	//Date: Thu, 10 Jan 2019 05:29:05 GMT
	//Content-Length: 65
	//
	//{"code":200,"msg":"success","data":{"Reply":"Thx for \"hello\""}}%
}

// If error is CodeError(github.com/xuxinx/cerr), will response the specific error code.
// Otherwise, return {"code":500,"msg":"system error"}
func Example_cerr() {
	http.Handle("/cerr", jsonh.ToHandler(func(_ *http.Request) error {
		return cerr.New(1001, "special error")
	}))

	log.Fatal(http.ListenAndServe(":8080", nil))

	// Output:
	//$ curl -i 'localhost:8080/cerr'
	//HTTP/1.1 400 Bad Request
	//Content-Type: application/json
	//Date: Thu, 10 Jan 2019 05:32:17 GMT
	//Content-Length: 35
	//
	//{"code":1001,"msg":"special error"}%
}
