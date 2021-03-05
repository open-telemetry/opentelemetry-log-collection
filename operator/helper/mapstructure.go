package helper

import (
	"encoding/json"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// make mapstructure use struct UnmarshalJSON to decode
func JSONUnmarshalerHook() mapstructure.DecodeHookFunc {
	return func(from reflect.Value, to reflect.Value) (interface{}, error) {
		if to.CanAddr() {
			to = to.Addr()
		}

		// If the destination implements the unmarshaling interface
		u, ok := to.Interface().(json.Unmarshaler)
		if !ok {
			return from.Interface(), nil
		}

		// If it is nil and a pointer, create and assign the target value first
		if to.IsNil() && to.Type().Kind() == reflect.Ptr {
			to.Set(reflect.New(to.Type().Elem()))
			u = to.Interface().(json.Unmarshaler)
		}
		v := from.Interface()
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		if err := u.UnmarshalJSON(bytes); err != nil {
			return to.Interface(), err
		}
		return to.Interface(), nil
	}
}
