package gorule

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

func modifyInterface(mod interface{}, tree []string, value string) error {

	t := reflect.TypeOf(mod)
	v := reflect.ValueOf(mod)

	log.Printf("t: %+v", t)
	log.Printf("v: %+v", v)

	var v2 reflect.Value
	var t2 reflect.Type
	if v.Kind() == reflect.Ptr {
		//t = reflect.TypeOf(reflect.Indirect(v))
		v2 = reflect.Indirect(v)
		t2 = v2.Type()
	} else {
		v2 = v
		t2 = t
	}

	for i := 0; i < t2.NumField(); i++ {
		field := t2.Field(i)
		log.Printf("field.Name: %s match: %s", field.Name, tree[0])
		if strings.EqualFold(field.Name, tree[0]) {

			/*if len(tree) > 1 {
				err := modifyInterface(t.Field(i), tree[1:], value)
				if err != nil {
					return err
				}
			} else {*/
			switch v.Elem().Field(i).Type().String() {
			case "int":
				newValue, err := strconv.Atoi(value)
				if err != nil {
					return err
				}
				v.Elem().Field(i).SetInt(int64(newValue))
			case "string":
				log.Printf("--- set string %s to %s", v.Elem().Field(i), value)
				v.Elem().Field(i).SetString(value)
			default:
				log.Printf("KIND: %+v", v.Elem().Field(i).Kind())
				switch v.Elem().Field(i).Kind() {
				case reflect.Ptr:
					//}
					//if v.Elem().Field(i).Kind() == reflect.Ptr {
					//if strings.HasPrefix(v.Elem().Field(i).Type().String(), "*") {
					return modifyInterface(v.Elem().Field(i).Interface(), tree[1:], value)
				case reflect.Struct:
					return modifyInterface(v.Elem().Field(i).Interface(), tree[1:], value)
				}
				return fmt.Errorf("type '%s' has not been implemented yet in the interface modifier", v.Elem().Field(i).Type().String())
			}

			//}
			continue
		}
	}
	// if we are a pointer, get the indirect type
	/*if v.Kind() == reflect.Ptr {
	  //t = reflect.TypeOf(reflect.Indirect(v))
	  v = reflect.Indirect(v)
	  t = v.Type()
	}*/
	/*
		case reflect.Float64:
		  attribute.S = aws.String(strconv.FormatFloat(v.FieldByName(field.Name).Float(), 'f', 10, 64))
		case reflect.Int:
		  attribute.N = aws.String(strconv.FormatInt(v.FieldByName(field.Name).Int(), 10))
		case reflect.Uint64:
		  attribute.N = aws.String(strconv.FormatUint(v.FieldByName(field.Name).Uint(), 10))
		case reflect.Int64:
		  attribute.N = aws.String(strconv.FormatInt(v.FieldByName(field.Name).Int(), 10))
		case reflect.Bool:
		  attribute.BOOL = aws.Bool(v.FieldByName(field.Name).Bool())
	*/
	return nil
}
