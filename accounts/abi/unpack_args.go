package abi

import (
	"fmt"
	"reflect"
)
// Unpack input in v according to the abi specification
func (abi ABI) UnpackArgs(v interface{}, method Method, input []byte) error {

	if len(input) == 0 {
		return fmt.Errorf("abi: unmarshalling empty input")
	}

	// make sure the passed value is a pointer
	valueOf := reflect.ValueOf(v)
	if reflect.Ptr != valueOf.Kind() {
		return fmt.Errorf("abi: Unpack(non-pointer %T)", v)
	}

	var (
		value = valueOf.Elem()
		typ   = value.Type()
	)

	switch value.Kind() {

	case reflect.Slice:
		if !value.Type().AssignableTo(r_interSlice) {
			return fmt.Errorf("abi: cannot marshal tuple in to slice %T (only []interface{} is supported)", v)
		}

		// if the slice already contains values, set those instead of the interface slice itself.
		if value.Len() > 0 {
			if len(method.Inputs) > value.Len() {
				return fmt.Errorf("abi: cannot marshal in to slices of unequal size (require: %v, got: %v)", len(method.Inputs), value.Len())
			}

			for i := 0; i < len(method.Inputs); i++ {
				marshalledValue, err := toGoType(i, method.Inputs[i], input)
				if err != nil {
					return err
				}
				reflectValue := reflect.ValueOf(marshalledValue)
				if err := set(value.Index(i).Elem(), reflectValue, method.Inputs[i]); err != nil {
					return err
				}
			}
			return nil
		}

		// create a new slice and start appending the unmarshalled
		// values to the new interface slice.
		z := reflect.MakeSlice(typ, 0, len(method.Inputs))
		for i := 0; i < len(method.Inputs); i++ {
			marshalledValue, err := toGoType(i, method.Inputs[i], input)
			if err != nil {
				return err
			}
			z = reflect.Append(z, reflect.ValueOf(marshalledValue))
		}
		value.Set(z)
	default:
		return fmt.Errorf("abi: cannot unmarshal tuple in to %v", typ)
	}


	return nil
}
