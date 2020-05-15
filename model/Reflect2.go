package DecentralizedABE

import (
	"encoding/json"
	"fmt"
	"github.com/Nik-U/pbc"
	"github.com/fatih/set"
	"math/big"
	"reflect"
	"strconv"
)

var specialHandle set.Interface
var curve CurveParam

func init() {
	specialHandle = set.New(set.NonThreadSafe)
	specialHandle.Add("*pbc.Params")
	specialHandle.Add("*pbc.Pairing")
	specialHandle.Add("*pbc.Element")
	specialHandle.Add("*big.Int")
	curve.Initialize()
}

func Serialize2Bytes(obj interface{}) ([]byte, error) {
	serialize2Map, err := Serialize2Map(obj)
	if err != nil {
		return nil, err
	}
	bytes, err := json.Marshal(serialize2Map)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func Serialize2Map(obj interface{}) (map[string]interface{}, error) {
	var err error
	if obj == nil {
		return nil, fmt.Errorf("nil input")
	}
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	data := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		if specialHandle.Has(field.Type.String()) {
			data[field.Name] = serializeHandle(field.Type, value)
			continue
		}
		switch field.Type.Kind() {
		case reflect.Array:
			tempArray := value.Interface().([]interface{})
			tempData := make([]interface{}, len(tempArray))
			for i, v := range tempArray {
				tempData[i], err = Serialize2Map(v)
				if err != nil {
					return nil, err
				}
			}
			data[field.Name] = tempData
			continue
		case reflect.Map:
			tempData := make(map[string]interface{}, len(value.MapKeys()))
			for _, key := range value.MapKeys() {
				innerVal := value.MapIndex(key)
				tempData[key.Interface().(string)], err = Serialize2Map(innerVal)
				if err != nil {
					return nil, err
				}
			}
			data[field.Name] = tempData
			continue
		case reflect.Struct:
			tempData, err := Serialize2Map(value)
			if err != nil {
				return nil, err
			}
			data[field.Name] = tempData
			continue
		default:
			data[field.Name] = value.Interface()
		}
	}
	return data, nil
}

func Deserialize2Struct(bytes []byte, obj interface{}) error {
	data := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &data); err != nil {
		fmt.Println(err.Error())
		return err
	}
	i, e := deserialize2Struct(data, obj)
	fmt.Println(i)
	return e
}

func deserialize2Struct(data map[string]interface{}, obj interface{}) (interface{}, error) {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		if specialHandle.Has(field.Type.String()) {
			result, err := deserializeHandle(field.Type, data[field.Name], field.Tag)
			if err != nil {
				return nil, err
			}
			value.Set(reflect.ValueOf(result))
		}

		switch field.Type.Kind() {
		case reflect.Array:
			innerType := field.Type.Elem()
			tempArray := data[field.Name].([]interface{})
			tempData := make([]interface{}, len(tempArray))
			if specialHandle.Has(innerType.String()) {
				for i, v := range tempArray {
					result, err := deserializeHandle(innerType, v, field.Tag)
					if err != nil {
						return nil, err
					}
					tempData[i] = result
				}
			} else {
				for i, v := range tempArray {
					result, err := deserialize2Struct(v.(map[string]interface{}), reflect.New(innerType))
					if err != nil {
						return nil, err
					}
					tempData[i] = result
				}
			}
			value.Set(reflect.ValueOf(tempData))
			continue
		case reflect.Map:
			innerType := field.Type.Elem()
			tempMap := data[field.Name].(map[string]interface{})
			//tempData := make(map[string]interface{}, len(tempMap))
			if specialHandle.Has(innerType.String()) {
				for k, v := range tempMap {
					result, err := deserializeHandle(innerType, v, field.Tag)
					if err != nil {
						return nil, err
					}
					value.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(result))
					//tempData[k] = result
				}
			} else {
				for k, v := range tempMap {
					result, err := deserialize2Struct(v.(map[string]interface{}), reflect.New(innerType))
					if err != nil {
						return nil, err
					}
					value.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(result))
					//tempData[k] = result
				}
			}
			//value.Set(reflect.ValueOf(tempData))
			continue
		case reflect.Struct:
			result, err := deserialize2Struct(data[field.Name].(map[string]interface{}), reflect.New(field.Type))
			if err != nil {
				return nil, err
			}
			value.Set(reflect.ValueOf(result))
			continue
		default:
			fmt.Println(field.Type.Kind())
			fmt.Println(field.Type.Name())
			fmt.Println(field.Type.String())
			value.Set(reflect.ValueOf(data[field.Name]))
		}
	}

	return obj, nil
}

func serializeHandle(fieldType reflect.Type, val reflect.Value) interface{} {
	switch fieldType.String() {
	case "*pbc.Params":
		return ""
	case "*pbc.Pairing":
		return ""
	case "*pbc.Element":
		return (val.Interface().(*pbc.Element)).String()
	case "*big.Int":
		return (val.Interface().(*big.Int)).String()
	default:
		return val.Interface()
	}
}

func deserializeHandle(fieldType reflect.Type, obj interface{}, tag reflect.StructTag) (interface{}, error) {
	if fieldType.Kind() == reflect.Struct {
		return deserialize2Struct(obj.(map[string]interface{}), reflect.New(fieldType))
	}
	switch fieldType.String() {
	case "*pbc.Params":
		return curve.Param, nil
	case "*pbc.Pairing":
		return curve.Pairing, nil
	case "*pbc.Element":
		fieldStr := tag.Get("field")
		fmt.Println(fieldType.Kind())
		fmt.Println(fieldType.Name())
		fmt.Println(fieldStr)
		i, err := strconv.Atoi(fieldStr)
		if err != nil {
			return nil, err
		}
		element, b := curve.Pairing.NewUncheckedElement(pbc.Field(i)).SetString(obj.(string), 10)
		if !b {
			return nil, fmt.Errorf("deserialze pbc.Element error with" + obj.(string))
		}
		return element, nil
	case "*big.Int":
		result, b := new(big.Int).SetString(obj.(string), 10)
		if !b {
			return nil, fmt.Errorf("deserialze big.Int error with" + obj.(string))
		}
		return result, nil
	default:
		return obj, nil
	}
}
