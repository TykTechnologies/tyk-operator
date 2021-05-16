// This file was taken from https://github.com/jinzhu/copier/blob/master/copier.go
package migrate

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

func copyAPI(a *v1alpha1.APIDefinitionSpec) v1alpha1.APIDefinitionSpec {
	o := v1alpha1.APIDefinitionSpec{
		Name: a.Name,
	}
	use(&o, a)
	return o
}

func copyPolicy(a *v1alpha1.SecurityPolicySpec) v1alpha1.SecurityPolicySpec {
	o := v1alpha1.SecurityPolicySpec{
		Name: a.Name,
	}
	use(&o, a)
	return o
}

func use(dest, src interface{}) {
	xr := reflect.ValueOf(dest)
	yr := reflect.ValueOf(src)
	if xr.Kind() == reflect.Ptr {
		xr = xr.Elem()
	}
	if yr.Kind() == reflect.Ptr {
		yr = yr.Elem()
	}
	switch yr.Kind() {
	case reflect.Struct:
		for i := 0; i < yr.NumField(); i++ {
			f := yr.Field(i)
			if zero(f) {
				continue
			}
			nf := xr.Field(i)
			switch f.Kind() {
			case reflect.Ptr:
				if f.Elem().Kind() == reflect.Struct {
					nf.Set(reflect.New(f.Elem().Type()))
					to := nf.Interface()
					from := f.Interface()
					use(to, from)
					continue
				}
			case reflect.Map:
				nf.Set(reflect.MakeMap(f.Type()))
			case reflect.Slice:
				nf.Set(reflect.MakeSlice(f.Type(), f.Len(), f.Cap()))
				to := nf.Interface()
				from := f.Interface()
				use(to, from)
			case reflect.Struct:
				to := nf.Addr().Interface()
				from := f.Addr().Interface()
				use(to, from)
			default:
				nf.Set(f)
			}
		}
	case reflect.Map:
		for _, key := range yr.MapKeys() {
			value := yr.MapIndex(key)
			if zero(value) {
				continue
			}
			kind := value.Kind()
			if kind == reflect.Struct {
				v := reflect.New(value.Type())
				use(v, value.Addr().Interface())

				// set the value
				xr.SetMapIndex(key, v.Elem())
			} else if kind == reflect.Ptr && value.Elem().Kind() == reflect.Struct {
				v := reflect.New(value.Type())
				use(v, value.Interface())
				// set the value
				xr.SetMapIndex(key, v)
			} else {
				xr.SetMapIndex(key, value)
			}
		}
	case reflect.Slice:
		for i := 0; i < yr.Len(); i++ {
			v := yr.Index(i)
			if !zero(v) {
				reflect.Append(xr, v)
			}
		}
	default:
		xr.Set(yr)
	}
}

func zero(v reflect.Value) bool {
	if v.IsZero() {
		return true
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		o := v.Interface()
		n := reflect.New(v.Type()).Elem().Interface()
		if jsonE(o, n) {
			return true
		}
		return fieldsE(v)
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if !zero(v.Index(i)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func jsonE(a, b interface{}) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return bytes.Equal(x, y)
}

func fieldsE(v reflect.Value) bool {
	for i := 0; i < v.NumField(); i++ {
		if !zero(v.Field(i)) {
			return false
		}
	}
	return true
}
