package matrix

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/dgruber/drmaa2interface"
	"github.com/mitchellh/copystructure"
)

// GetNextValue returns the next value of the given number encoded
// as a slice of ints where each int represents a position in a number.
// The number is incremented by one and the result is returned.
// In case of an overflow an error is returned.
// maxNumber defines the max. value of each position in the number,
// a decimal number with 3 digits is encoded as []int{9, 9, 9}.
func GetNextValue(maxNumber []int, currentNumber []int) ([]int, error) {
	if len(maxNumber) != len(currentNumber) {
		return nil, fmt.Errorf("maxNumber and currentNumber have a different length")
	}
	for i := 0; i < len(maxNumber); i++ {
		if maxNumber[i] < currentNumber[i] {
			return nil, fmt.Errorf("maxNumber[%d] < currentNumber[%d]", i, i)
		}
	}
	for i := len(currentNumber) - 1; i >= 0; i-- {
		currentNumber[i] += 1
		if currentNumber[i] > maxNumber[i] {
			currentNumber[i] = 0
			// overflow, add +1 to previous number
			continue
		}
		return currentNumber, nil
	}
	return nil, fmt.Errorf("overflow")
}

func CopyJobTemplate(jt drmaa2interface.JobTemplate) (drmaa2interface.JobTemplate, error) {
	copy, err := copystructure.Copy(jt)
	out := copy.(drmaa2interface.JobTemplate)
	return out, err
}

func ReplaceInField(jt drmaa2interface.JobTemplate, fieldName, pattern, replacement string) (drmaa2interface.JobTemplate, error) {

	copyJT, err := CopyJobTemplate(jt)
	if err != nil {
		return jt, fmt.Errorf("error copying job template: %s", err)
	}

	jtValue := reflect.ValueOf(&copyJT)

	jtField := reflect.Indirect(jtValue).FieldByName(fieldName)
	if !jtField.IsValid() {
		return jt, fmt.Errorf("unknown JobTemplate field name %s", fieldName)
	}

	switch jtField.Kind() {
	case reflect.String:
		// for string fields replace each occurrence of the pattern (RemoteCommand: app_{{1}} -> app_X)
		v := strings.Replace(jtField.String(), pattern, replacement, -1)
		jtField.SetString(v)
	case reflect.Bool:
		// for bool fields set to true or false deping if replacements is a bool value
		bV, err := strconv.ParseBool(replacement)
		if err != nil {
			return jt, fmt.Errorf("replacement %s for field %s is not a bool value: %v",
				replacement, fieldName, err)
		}
		jtField.SetBool(bV)
	case reflect.Int32, reflect.Int64:
		// for int fields set to the int value of the replacement
		i, err := strconv.Atoi(replacement)
		if err != nil {
			return jt, fmt.Errorf("failed to convert %s to int: %s", jtField.String(), err)
		}
		jtField.SetInt(int64(i))
	case reflect.Slice:
		// for slice fields replace each occurrence of the pattern
		if jtField.Len() > 0 {
			if jtField.Index(0).Kind() != reflect.String {
				return jt, fmt.Errorf("field %s is not a string slice", fieldName)
			}
			v := make([]string, 0)
			for _, s := range jtField.Interface().([]string) {
				v = append(v, strings.Replace(s, pattern, replacement, -1))
			}
			jtField.Set(reflect.ValueOf(v))
		}
	case reflect.Map:
		// for map fields replace each occurrence of the pattern found in any key
		// or value. Expects a map of strings.
		iter := jtField.MapRange()
		m := map[string]string{}
		for iter.Next() {
			if iter.Key().Kind() != reflect.String || iter.Value().Kind() != reflect.String {
				return jt, fmt.Errorf("unsupported map type %s or value %s for field %s",
					iter.Key().Type(), iter.Value().Type(), fieldName)
			}
			// key and value are strings - try replacing pattern in key and value
			k := strings.Replace(iter.Key().String(), pattern, replacement, -1)
			v := strings.Replace(iter.Value().String(), pattern, replacement, -1)
			m[k] = v
		}
		jtField.Set(reflect.ValueOf(m))
	default:
		return jt, fmt.Errorf("unsupported field type %s", jtField.Kind())
	}
	return copyJT, nil
}
