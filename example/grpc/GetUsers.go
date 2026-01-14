package grpc

import (
	context "context"
	"fmt"
)

type API struct {
	UnimplementedAPIServer
}

func (API) GetUsers(context.Context, *Null) (*Null, error) {
	fmt.Println(123)
	return &Null{}, nil
}
