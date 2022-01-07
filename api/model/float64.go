package model

import (
	"encoding/json"
	"fmt"
	"sync"
)

//Float64 is a work around to allow representing floating points as strings
// +kubebuilder:validation:Pattern="^(?:[-+]?(?:[0-9]+))?(?:\\.[0-9]*)?(?:[eE][\\+\\-]?(?:[0-9]+))?$"
//
// source for pattern https://github.com/asaskevich/govalidator/blob/7a23bdc65eef5f3783e782b436f3065eae3fc72d/patterns.go#L19
type Float64 string

// number when this is true we marshal Float64 as a number
var number bool
var numMu sync.Mutex

// MarshalJSON marshalls f as a float64 number
func (f Float64) MarshalJSON() ([]byte, error) {
	if number {
		v, err := json.Number(string(f)).Float64()
		if err != nil {
			return nil, err
		}

		return json.Marshal(v)
	}

	return json.Marshal(string(f))
}

// UnmarshalJSON supports both float64 and a json.Number of float64
func (f *Float64) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch e := v.(type) {
	case float64:
		*f = Float64(fmt.Sprint(e))
		return nil
	case string:
		_, err := json.Number(e).Float64()
		if err != nil {
			return err
		}

		*f = Float64(e)

		return nil
	default:
		return fmt.Errorf("operator: failed to decode type %#T to a Float64", e)
	}
}

// Marshal marshals v as json. This makes sure Float64 values are treated as a number of float64 type
func Marshal(v interface{}) ([]byte, error) {
	numMu.Lock()
	number = true

	defer func() {
		number = false
		numMu.Unlock()
	}()

	return json.Marshal(v)
}

// Percent describes a percentage value expressed as a float. This is a positive
// decimal value that is less than 1
//+kubebuilder:validation:Pattern="^0\\.\\d+|1\\.0$"
type Percent string

// MarshalJSON returns a json string for p. This is a string for normal
// operations and a float64 when marshalling for dashboard or gateway
func (p Percent) MarshalJSON() ([]byte, error) {
	return Float64(p).MarshalJSON()
}

func (f *Percent) UnmarshalJSON(b []byte) error {
	var x Float64

	err := x.UnmarshalJSON(b)
	if err != nil {
		return err
	}

	*f = Percent(x)

	return nil
}
