package funcutils

import (
	"context"
	"errors"
	"fmt"
	"go/token"
	"log"
	"reflect"

	"github.com/wuqunyong/file_storage/pkg/errs"
)

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()
var typeOfCodeError = reflect.TypeOf((*errs.CodeError)(nil)).Elem()
var typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}

type MethodType struct {
	Signature string
	Type      reflect.Type
	Func      reflect.Value
	ArgType   []reflect.Type
	ReplyType reflect.Type
	NumIn     int
}

func GetRPCReflectFunc(fun any, reportErr bool) *MethodType {
	mtype := reflect.TypeOf(fun)
	if mtype.Kind() != reflect.Func {
		return nil
	}

	mname := mtype.String()

	switch mtype.NumIn() {
	// context.Context, request, response
	// func (actor *ActorObjA) Func2(ctx context.Context, arg *testdata.Person, reply *testdata.Person) errs.CodeError {
	case 3:
		iNumIn := 3
		// Method needs two ins: context.Context, *args, *reply.
		if mtype.NumIn() != iNumIn {
			if reportErr {
				log.Printf("rpc.Register: method %q has %d input parameters; needs exactly three", mname, mtype.NumIn())
			}
			return nil
		}

		var inArgs []reflect.Type
		for i := 0; i < mtype.NumIn(); i++ {
			t := mtype.In(i)
			inArgs = append(inArgs, t)
		}

		// First arg must be a pointer.
		contextType := mtype.In(0)
		if contextType != typeOfContext {
			if reportErr {
				log.Printf("rpc.Register: context.Context of method %q is not a context.Context: %q", mname, contextType)
			}
			return nil
		}

		argType := mtype.In(1)
		if argType.Kind() != reflect.Ptr {
			if reportErr {
				log.Printf("rpc.Register: argument type of method %q is not a pointer: %q", mname, argType)
			}
			return nil
		}
		if !isExportedOrBuiltinType(argType) {
			if reportErr {
				log.Printf("rpc.Register: argument type of method %q is not exported: %q", mname, argType)
			}
			return nil
		}
		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			if reportErr {
				log.Printf("rpc.Register: reply type of method %q is not a pointer: %q", mname, replyType)
			}
			return nil
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			if reportErr {
				log.Printf("rpc.Register: reply type of method %q is not exported: %q", mname, replyType)
			}
			return nil
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			if reportErr {
				log.Printf("rpc.Register: method %q has %d output parameters; needs exactly one", mname, mtype.NumOut())
			}
			return nil
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfCodeError {
			if reportErr {
				log.Printf("rpc.Register: return type of method %q is %q, must be CodeError", mname, returnType)
			}
			return nil
		}

		return &MethodType{
			Signature: mname,
			Type:      mtype,
			Func:      reflect.ValueOf(fun),
			ArgType:   inArgs,
			ReplyType: replyType,
			NumIn:     iNumIn,
		}
		// context.Context, notify
		// func (actor *ActorObjA) Func3(ctx context.Context, arg *testdata.MSG_NOTICE_INSTANCE) {
	case 2:
		iNumIn := 2
		var inArgs []reflect.Type
		for i := 0; i < mtype.NumIn(); i++ {
			t := mtype.In(i)
			inArgs = append(inArgs, t)
		}

		// First arg must be a pointer.
		contextType := mtype.In(0)
		if contextType != typeOfContext {
			if reportErr {
				log.Printf("rpc.Register: context.Context of method %q is not a context.Context: %q", mname, contextType)
			}
			return nil
		}

		argType := mtype.In(1)
		if argType.Kind() != reflect.Ptr {
			if reportErr {
				log.Printf("rpc.Register: argument type of method %q is not a pointer: %q", mname, argType)
			}
			return nil
		}
		if !isExportedOrBuiltinType(argType) {
			if reportErr {
				log.Printf("rpc.Register: argument type of method %q is not exported: %q", mname, argType)
			}
			return nil
		}

		// Method needs no return
		if mtype.NumOut() != 0 {
			if reportErr {
				log.Printf("rpc.Register: method %q has %d output parameters; needs exactly zero", mname, mtype.NumOut())
			}
			return nil
		}

		return &MethodType{
			Signature: mname,
			Type:      mtype,
			Func:      reflect.ValueOf(fun),
			ArgType:   inArgs,
			NumIn:     iNumIn,
		}

	default:
		if reportErr {
			log.Printf("rpc.Register: method %q has %d input parameters; needs three or two", mname, mtype.NumIn())
		}
		return nil
	}
}

func CallPRCReflectRequestFunc[T1 any, T2 any](method *MethodType, ctx context.Context, arg1 T1, arg2 T2) (errs.CodeError, error) {
	if reflect.TypeOf(arg1) != method.ArgType[1] {
		return nil, errors.New("invalid arg1")
	}

	if reflect.TypeOf(arg2) != method.ReplyType {
		return nil, errors.New("invalid arg2")
	}

	args := make([]reflect.Value, method.NumIn)
	args[0] = reflect.ValueOf(ctx)
	args[1] = reflect.ValueOf(arg1)
	args[2] = reflect.ValueOf(arg2)

	rets := method.Func.Call(args)
	if len(rets) != 1 {
		return nil, errors.New("invalid rets len")
	}

	var value any
	if value = rets[0].Interface(); value == nil {
		return nil, nil
	}

	result, ok := value.(errs.CodeError)
	if !ok {
		return nil, errors.New("invalid rets type err")
	}
	return result, nil
}

func CallPRCReflectNotifyFunc[T1 any](method *MethodType, ctx context.Context, arg1 T1) error {
	if reflect.TypeOf(arg1) != method.ArgType[1] {
		return errors.New("invalid arg1")
	}

	args := make([]reflect.Value, method.NumIn)
	args[0] = reflect.ValueOf(ctx)
	args[1] = reflect.ValueOf(arg1)

	rets := method.Func.Call(args)
	if len(rets) != 0 {
		return errors.New("invalid rets len")
	}

	return nil
}

// func(client *Client, request *PB, response *PB) errs.CodeError
func GetClientReflectFunc(fun any) (*MethodType, error) {
	mtype := reflect.TypeOf(fun)
	if mtype.Kind() != reflect.Func {
		return nil, errors.New("handler needs to be a func")
	}

	mname := mtype.String()

	iNumIn := 3
	// Method needs three ins: *Client, *args, *reply.
	if mtype.NumIn() != iNumIn {
		return nil, fmt.Errorf("rpc.Register: method %q has %d input parameters; needs exactly  3", mname, mtype.NumIn())
	}

	var inArgs []reflect.Type
	for i := 0; i < mtype.NumIn(); i++ {
		t := mtype.In(i)
		inArgs = append(inArgs, t)
	}

	// First arg must be a pointer.
	argType := mtype.In(1)
	if argType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("rpc.Register: argument type of method %q is not a pointer: %q", mname, argType)
	}
	if !isExportedOrBuiltinType(argType) {
		return nil, fmt.Errorf("rpc.Register: argument type of method %q is not exported: %q", mname, argType)
	}
	// Second arg must be a pointer.
	replyType := mtype.In(2)
	if replyType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("rpc.Register: reply type of method %q is not a pointer: %q", mname, replyType)
	}
	// Reply type must be exported.
	if !isExportedOrBuiltinType(replyType) {
		return nil, fmt.Errorf("rpc.Register: reply type of method %q is not exported: %q", mname, replyType)
	}
	// Method needs one out.
	if mtype.NumOut() != 1 {
		return nil, fmt.Errorf("rpc.Register: method %q has %d output parameters; needs exactly one", mname, mtype.NumOut())
	}
	// The return type of the method must be error.
	if returnType := mtype.Out(0); returnType != typeOfCodeError {
		return nil, fmt.Errorf("rpc.Register: return type of method %q is %q, must be CodeError", mname, returnType)
	}

	return &MethodType{
		Signature: mname,
		Type:      mtype,
		Func:      reflect.ValueOf(fun),
		ArgType:   inArgs,
		ReplyType: replyType,
		NumIn:     iNumIn,
	}, nil
}

func CallClientReflectFunc[T1 any, T2 any](method *MethodType, client any, arg1 T1, arg2 T2) (errs.CodeError, error) {
	if reflect.TypeOf(arg1) != method.ArgType[1] {
		return nil, errors.New("invalid arg1")
	}

	if reflect.TypeOf(arg2) != method.ReplyType {
		return nil, errors.New("invalid arg2")
	}

	args := make([]reflect.Value, method.NumIn)
	args[0] = reflect.ValueOf(client)
	args[1] = reflect.ValueOf(arg1)
	args[2] = reflect.ValueOf(arg2)

	rets := method.Func.Call(args)
	if len(rets) != 1 {
		return nil, errors.New("invalid rets len")
	}

	var value any
	if value = rets[0].Interface(); value == nil {
		return nil, nil
	}

	result, ok := value.(errs.CodeError)
	if !ok {
		return nil, errors.New("invalid rets type err")
	}
	return result, nil
}
