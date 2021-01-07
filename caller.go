package socketio

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

type caller struct {
	Func       reflect.Value
	Args       []reflect.Type
	NeedSocket bool
}

func newCaller(f interface{}) (*caller, error) {
	fv := reflect.ValueOf(f)
	if fv.Kind() != reflect.Func {
		return nil, fmt.Errorf("f is not func")
	}
	ft := fv.Type()
	if ft.NumIn() == 0 {
		return &caller{
			Func: fv,
		}, nil
	}
	args := make([]reflect.Type, ft.NumIn())
	for i, n := 0, ft.NumIn(); i < n; i++ {
		args[i] = ft.In(i)
	}
	needSocket := false
	if args[0].Name() == "Socket" {
		args = args[1:]
		needSocket = true
	}
	return &caller{
		Func:       fv,
		Args:       args,
		NeedSocket: needSocket,
	}, nil
}

func (c *caller) GetArgs() []interface{} {
	ret := make([]interface{}, len(c.Args))
	for i, argT := range c.Args {
		if argT.Kind() == reflect.Ptr {
			argT = argT.Elem()
		}
		v := reflect.New(argT)
		ret[i] = v.Interface()
	}
	return ret
}

func (c *caller) Call(so Socket, eventName string, args []interface{}, unhandledTrigger bool) []reflect.Value {
	var a []reflect.Value
	diff := 0
	//if unhandledTrigger {  //cannot change data length
	//	diff += 1
	//}
	if c.NeedSocket {
		diff += 1
		a = make([]reflect.Value, len(args)+diff)
		a[0] = reflect.ValueOf(so)

		if unhandledTrigger {
			diff += 1
			a[1] = reflect.ValueOf(eventName)
			//a = append(a, reflect.ValueOf(eventName))
		}
	} else {
		a = make([]reflect.Value, len(args)+diff)

		if unhandledTrigger {
			diff += 1
			a[0] = reflect.ValueOf(eventName)
			//a = append(a, reflect.ValueOf(eventName))
		}
	}

	if len(args) != len(c.Args) {
		log.Println("args not match event register function by:", eventName)
		return []reflect.Value{reflect.ValueOf([]interface{}{}), reflect.ValueOf(errors.New("Arguments do not match"))}
	}

	for i, arg := range args {
		v := reflect.ValueOf(arg)
		if c.Args[i].Kind() != reflect.Ptr {
			if v.IsValid() {
				v = v.Elem()
			} else {
				v = reflect.Zero(c.Args[i])
			}
		}
		if len(a) > i+diff {
			a[i+diff] = v
		}
	}

	return c.Func.Call(a)
}
