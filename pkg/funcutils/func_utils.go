package funcutils

import (
	"context"
	"errors"
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
	ArgType   reflect.Type
	ReplyType reflect.Type
	NumIn     int
}

func GetReflectFunc(fun any, reportErr bool) *MethodType {
	mtype := reflect.TypeOf(fun)
	if mtype.Kind() != reflect.Func {
		return nil
	}

	mname := mtype.String()

	iNumIn := 3
	// Method needs two ins: context.Context, *args, *reply.
	if mtype.NumIn() != iNumIn {
		if reportErr {
			log.Printf("rpc.Register: method %q has %d input parameters; needs exactly two\n", mname, mtype.NumIn())
		}
		return nil
	}
	// First arg must be a pointer.
	contextType := mtype.In(0)
	if contextType != typeOfContext {
		if reportErr {
			log.Printf("rpc.Register: context.Context of method %q is not a context.Context: %q\n", mname, contextType)
		}
		return nil
	}

	argType := mtype.In(1)
	if argType.Kind() != reflect.Ptr {
		if reportErr {
			log.Printf("rpc.Register: argument type of method %q is not a pointer: %q\n", mname, argType)
		}
		return nil
	}
	if !isExportedOrBuiltinType(argType) {
		if reportErr {
			log.Printf("rpc.Register: argument type of method %q is not exported: %q\n", mname, argType)
		}
		return nil
	}
	// Second arg must be a pointer.
	replyType := mtype.In(2)
	if replyType.Kind() != reflect.Ptr {
		if reportErr {
			log.Printf("rpc.Register: reply type of method %q is not a pointer: %q\n", mname, replyType)
		}
		return nil
	}
	// Reply type must be exported.
	if !isExportedOrBuiltinType(replyType) {
		if reportErr {
			log.Printf("rpc.Register: reply type of method %q is not exported: %q\n", mname, replyType)
		}
		return nil
	}
	// Method needs one out.
	if mtype.NumOut() != 1 {
		if reportErr {
			log.Printf("rpc.Register: method %q has %d output parameters; needs exactly one\n", mname, mtype.NumOut())
		}
		return nil
	}
	// The return type of the method must be error.
	if returnType := mtype.Out(0); returnType != typeOfCodeError {
		if reportErr {
			log.Printf("rpc.Register: return type of method %q is %q, must be CodeError\n", mname, returnType)
		}
		return nil
	}

	return &MethodType{
		Signature: mname,
		Type:      mtype,
		Func:      reflect.ValueOf(fun),
		ArgType:   argType,
		ReplyType: replyType,
		NumIn:     iNumIn,
	}
}

func CallReflectFunc[T1 any, T2 any](method *MethodType, ctx context.Context, arg1 T1, arg2 T2) (errs.CodeError, error) {
	if reflect.TypeOf(arg1) != method.ArgType {
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
