# jsonh
[![GoDoc](https://godoc.org/github.com/xuxinx/jsonh?status.svg)](https://godoc.org/github.com/xuxinx/jsonh)  
Convert Go func to http.Handler that handle json request and response json.  
Write http json api easily and quickly !  

## Examples
#### request and response
```go
  http.Handle("/hello", jsonh.ToHandler(func() error {
  	return nil
  }))
  // $ curl -i localhost:8080/hello
  // HTTP/1.1 200 OK
  // Content-Type: application/json
  //
  // {"code":200,"msg":"success"}
  
  // input and output are optional
  type Input struct {
  	Name string
  }
  type Output struct {
  	Greet string
  }
  http.Handle("/hello", jsonh.ToHandler(func(in *Input) (*Output, error) {
  	return &Output{Greet: fmt.Sprintf("Hello %s", in.Name)}, nil
  }))
  // $ curl -i -d '{"Name":"Tom"}' localhost:8889/hello
  // HTTP/1.1 200 OK
  // Content-Type: application/json
  //
  // {"code":200,"msg":"success","data":{"Greet":"Hello Tom"}}

  // output can be `any` type
  type Input struct {
  	Name string
  }
  http.Handle("/hello", jsonh.ToHandler(func(in *Input) (string, error) {
  	return fmt.Sprintf("Hello %s", in.Name), nil
  }))
  // $ curl -i -d '{"Name":"Tom"}' localhost:8889/hello
  // HTTP/1.1 200 OK
  // Content-Type: application/json
  //
  // {"code":200,"msg":"success","data":"Hello Tom"}
  
  // http.ResponseWriter and *http.Request are optional
  http.Handle("/hello", jsonh.ToHandler(func(wr http.ResponseWriter, r *http.Request, in *Input) (string, error) {
  	return fmt.Sprintf("Hello %s", in.Name), nil
  }))
```
#### response error
``` go
  // unknown error
  http.Handle("/hello", jsonh.ToHandler(func() error {
  	return errors.New("oh no")
  }))
  // $ curl -i localhost:8080/hello
  // HTTP/1.1 500 Internal Server Error
  // Content-Type: application/json
  //
  // {"code":500,"msg":"system error"}
  
  // specific error
  // using "github.com/xuxinx/jsonh/cerr"
  http.Handle("/hello", jsonh.ToHandler(func() error {
  	return cerr.New(1001, "this is error 1001")
  }))
  // $ curl -i localhost:8080/hello
  // HTTP/1.1 400 Bad Request
  // Content-Type: application/json
  //
  // {"code":1001,"msg":"this is error 1001"}
```
