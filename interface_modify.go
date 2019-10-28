package gorule

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var uint8slice = "[]uint8"

// modifyInterface gets the value of an interface based on tree
func modifyInterface(mod interface{}, tree []string, value string) error {
	_, _, v2, _ := getReflection(mod)
	//log.Printf("modifyInterface mod:%T type:%+v tree:%v value:%s", v2.Interface(), v2.Kind(), tree, value)

	if len(tree) == 0 {
		modifyValue(v2, tree, value)
	}

	switch v2.Kind() {
	case reflect.Struct:
		return modifyInterfaceStruct(mod, tree, value)
	case reflect.Map:
		return modifyInterfaceMap(mod, tree, value)
	case reflect.Slice:
		return modifyInterfaceSlice(mod, tree, value)
	default:
		return modifyValue(v2, tree, value)
	}
}

// modifyInterface gets the value of an interface based on tree
func modifyValue(v reflect.Value, tree []string, value string) error {
	var v2 reflect.Value

	// convert pointer to non-pointer
	if v.Kind() == reflect.Ptr {
		v2 = reflect.Indirect(v)
	} else {
		v2 = v
	}
	//log.Printf("modifyValue mod:%T type:%+v tree:%v value:%s", v2.Interface(), v2.Kind(), tree, value)

	switch v2.Kind() {
	case reflect.String:
		v2.SetString(value)
		return nil
	case reflect.Int:
		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("failed to convert '%s' to int: %s", value, err)
		}
		v2.SetInt(int64(i))
		return nil
	case reflect.Int64:
		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("failed to convert '%s' to int: %s", value, err)
		}
		v2.SetInt(int64(i))
		return nil
	case reflect.Bool:
		if strings.EqualFold(value, "true") {
			v2.SetBool(true)
		} else {
			v2.SetBool(false)
		}
		return nil
	case reflect.Struct:
		return modifyInterfaceStruct(v.Interface(), tree, value)
	case reflect.Map:
		return modifyInterfaceMap(v.Interface(), tree, value)
	case reflect.Slice:
		//log.Printf("setting slice of: %T", v.Interface())
		//log.Printf("setting slice of: %s", v.Kind())
		switch fmt.Sprintf("%T", v.Interface()) {
		case uint8slice: // []byte
			b := []uint8(value)
			v2.Set(reflect.ValueOf(b))
			return nil
		default:
			return modifyInterfaceSlice(v.Interface(), tree, value)
		}
	default:
		return fmt.Errorf("modifyValue type '%s' has not been found in the resource '%T'", tree[0:], v.Interface())
	}
}

// modifyInterfaceStruct gets the value of an interface based on tree of a Structure
func modifyInterfaceStruct(mod interface{}, tree []string, value string) error {
	v, _, v2, t2 := getReflection(mod)
	//log.Printf("modifyInterfaceStruct mod:%T type:%+v tree:%v value:%s", v2.Interface(), v2.Kind(), tree, value)
	// Loop through all field of the structure
	for i := 0; i < t2.NumField(); i++ {
		field := t2.Field(i)
		//log.Printf("modifyInterfaceStruct matching %s with %s", field.Name, tree[0])
		if strings.EqualFold(field.Name, tree[0]) {

			switch v.Elem().Field(i).Kind() {
			case reflect.Map, reflect.Slice, reflect.Ptr:
				// we only create a new element of this type if they are zero
				if isEmpty(v.Elem().Field(i)) {
					modNew, err := createStruct(v.Elem().Field(i))
					if err != nil {
						return fmt.Errorf("modifyInterfaceStruct failed to create instance for empty struct: %s", err)
					}
					//log.Printf("createStruct: set %T to %+v", v.Elem().Field(i).Interface(), modNew)
					v.Elem().Field(i).Set(reflect.ValueOf(modNew))
				}

			}

			return modifyValue(v.Elem().Field(i), tree[1:], value)
		}
	}
	return fmt.Errorf("modifyInterfaceStruct type '%s' has not been found in the resource '%T'", tree[0], v2.Interface())
}

// modifyInterfaceMap gets the value of an interface based on tree of a Map
func modifyInterfaceMap(mod interface{}, tree []string, value string) error {
	v, t, v2, _ := getReflection(mod)
	//log.Printf("modifyInterfaceMap mod:%T type:%+v tree:%v value:%s", v2.Interface(), v2.Kind(), tree, value)

	// Loop through all field of the structure
	for _, i := range v2.MapKeys() {
		//log.Printf("map match %s, %s", v2.MapIndex(i).Interface(), tree[0])
		if strings.EqualFold(i.String(), tree[0]) {
			return modifyValue(v2.MapIndex(i), tree[1:], value)
		}
	}
	// if element not found in MAP, then we add it
	// get the kind of the map key
	//log.Printf("map key kind is: %s", t.Key().Kind())
	switch t.Key().Kind() {
	case reflect.String:
		// get the kind of the map value
		//log.Printf("map value kind is: %s", t.Elem())
		switch t.Elem().Kind() {
		case reflect.Slice:
			//log.Printf("map int is: %+v, %+v LEN:%d", t, t, v.Len())
			// adding key to exising map
			//log.Printf("adding key to existing map of len:%d", v.Len())

			e := reflect.Indirect(reflect.New(t.Elem())).Interface()
			switch e.(type) {
			case []string:
				e = append(e.([]string), value)
			default:
				return fmt.Errorf("unknown type: %T", e)
			}
			v.SetMapIndex(reflect.ValueOf(tree[0]), reflect.ValueOf(e))
			return nil

		}
	}

	return fmt.Errorf("modifyInterfaceMap type '%s' has not been found in the resource '%T'", tree[0], v2.Interface())
}

// modifyInterfaceSlice gets the value of an interface based on tree of a Slice
func modifyInterfaceSlice(mod interface{}, tree []string, value string) error {
	_, _, v2, _ := getReflection(mod)
	//log.Printf("modifyInterfaceSlice mod:%T type:%+v tree:%v value:%s", v2.Interface(), v2.Kind(), tree, value)

	// Loop through all field of the structure
	if len(tree) == 0 {
		tree = append(tree, "0")
	}
	treeInt, err := strconv.Atoi(tree[0])
	if err != nil {
		return fmt.Errorf("modifyInterfaceSlice failed to convert '%s' in to a number: %s", tree[0], err)
	}

	for i := 0; i < v2.Len(); i++ {
		if i == treeInt {
			switch v2.Index(i).Kind() {
			case reflect.Ptr: // reflect.Map, reflect.Slice,
				// convert pointer to non-pointer
				//v3 := reflect.Indirect(v2)
				//log.Printf("got kind: %s", v2.Index(i).Kind())
				//log.Printf("got kind2: %T", v2.Index(i).Interface())
				return modifyInterface(v2.Index(i).Interface(), tree[1:], value)
			}
			return modifyValue(v2.Index(i), tree[1:], value)
		}
	}
	return fmt.Errorf("modifyInterfaceSlice slice '%s' has not been found in the resource '%T'", tree[0], v2.Interface())
}
