package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	coprocess "github.com/TykTechnologies/tyk-protobuf/bindings/go"
	"google.golang.org/grpc"
)

const (
	listenAddress = ":9999"
)

var (
	out = os.Stdout
)

func main() {
	lis, err := net.Listen("tcp", listenAddress)
	fatalOnErr(err)

	logInfo(fmt.Sprintf("starting gRPC server on %s", listenAddress))
	s := grpc.NewServer()
	coprocess.RegisterDispatcherServer(s, &dispatcher{})
	fatalOnErr(s.Serve(lis))
}

type dispatcher struct{}

func (d *dispatcher) Dispatch(ctx context.Context, object *coprocess.Object) (*coprocess.Object, error) {
	switch object.HookType {
	case coprocess.HookType_Pre:
		object.Request.SetHeaders = make(map[string]string)
		object.Request.SetHeaders["Pre"] = object.HookName

		return object, nil
	case coprocess.HookType_CustomKeyCheck:
		object.Request.SetHeaders = make(map[string]string)
		object.Request.SetHeaders["Auth"] = object.HookName

		headers := object.Request.GetHeaders()
		authorizationHeader, ok := headers["Authorization"]
		if !ok {
			object.Request.ReturnOverrides.ResponseError = "missing auth header"
			object.Request.ReturnOverrides.ResponseCode = http.StatusUnauthorized
			object.Request.ReturnOverrides.Headers = make(map[string]string)
			object.Request.SetHeaders["Auth"] = object.HookName

			return object, nil
		}

		if authorizationHeader != "foobarbaz" {
			object.Request.ReturnOverrides.ResponseError = "wrong auth header"
			object.Request.ReturnOverrides.ResponseCode = http.StatusUnauthorized
			object.Request.ReturnOverrides.Headers = make(map[string]string)
			object.Request.SetHeaders["Auth"] = object.HookName

			return object, nil
		}

		// artificial delay to test ID extractor
		time.Sleep(time.Second * 2)

		// Set the ID extractor deadline, useful for caching valid keys:
		extractorDeadline := time.Now().Add(time.Minute).Unix()

		object.Session = &coprocess.SessionState{
			Rate:                -1,
			Per:                 -1,
			QuotaMax:            -1,
			QuotaRenews:         time.Now().Unix(),
			IdExtractorDeadline: extractorDeadline,
		}

		object.Metadata = map[string]string{
			"token": authorizationHeader,
		}

		return object, nil
	case coprocess.HookType_PostKeyAuth:
		object.Request.SetHeaders = make(map[string]string)
		object.Request.SetHeaders["PostKeyAuth"] = object.HookName

		return object, nil
	case coprocess.HookType_Post:
		object.Request.SetHeaders = make(map[string]string)
		object.Request.SetHeaders["Post"] = object.HookName

		return object, nil
	}

	return object, fmt.Errorf("unknown hook: %s", object.HookName)
}

func (d *dispatcher) DispatchEvent(ctx context.Context, event *coprocess.Event) (*coprocess.EventReply, error) {
	return &coprocess.EventReply{}, nil
}

func logInfo(msg string) {
	_, _ = fmt.Fprintf(out, "INFO: %s\n", msg)
}

func fatalOnErr(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(out, "FATAL: %s\n", err.Error())
		os.Exit(1)
	}
}
