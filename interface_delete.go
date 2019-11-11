package gorule

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var uint8slice = "[]uint8"

// deleteInterface gets the value of an interface based on tree
func deleteInterface(mod interface{}, tree []string) error {
	_, _, v2, _ := getReflection(mod)
	//log.Printf("deleteInterface mod:%T type:%+v tree:%v value:%s", v2.Interface(), v2.Kind(), tree, value)

	if len(tree) == 0 {
		deleteValue(v2, tree)
	}

	switch v2.Kind() {
	case reflect.Struct:
		return deleteInterfaceStruct(mod, tree)
	case reflect.Map:
		return deleteInterfaceMap(mod, tree)
	case reflect.Slice:
		return deleteInterfaceSlice(mod, tree)
	default:
		return deleteValue(v2, tree)
	}
}

// deleteInterface gets the value of an interface based on tree
func deleteValue(v reflect.Value, tree []string) error {
	var v2 reflect.Value

	// convert pointer to non-pointer
	if v.Kind() == reflect.Ptr {
		v2 = reflect.Indirect(v)
	} else {
		v2 = v
	}
	//log.Printf("deleteValue mod:%T type:%+v tree:%v", v2.Interface(), v2.Kind(), tree)

	switch v2.Kind() {
	case reflect.String:
		v2.SetString("")
		return nil
	case reflect.Int:
		v2.SetInt(int64(0))
		return nil
	case reflect.Int64:
		v2.SetInt(int64(0))
		return nil
	case reflect.Bool:
		v2.SetBool(false)
		return nil
	case reflect.Struct:
		return deleteInterfaceStruct(v.Interface(), tree)
	case reflect.Map:
		return deleteInterfaceMap(v.Interface(), tree)
	case reflect.Slice:
		//log.Printf("setting slice of: %T", v.Interface())
		//log.Printf("setting slice of: %s", v.Kind())
		switch fmt.Sprintf("%T", v.Interface()) {
		case uint8slice: // []byte
			b := []uint8{}
			v2.Set(reflect.ValueOf(b))
			return nil
		default:
			return deleteInterfaceSlice(v.Interface(), tree)
		}
	default:
		return fmt.Errorf("deleteValue type '%s' has not been found in the resource '%T'", tree[0:], v.Interface())
	}
}

// deleteInterfaceStruct gets the value of an interface based on tree of a Structure
func deleteInterfaceStruct(mod interface{}, tree []string) error {
	v, _, _, t2 := getReflection(mod)
	//log.Printf("deleteInterfaceStruct mod:%T type:%+v tree:%v ", v2.Interface(), v2.Kind(), tree)
	// Loop through all field of the structure
	for i := 0; i < t2.NumField(); i++ {
		field := t2.Field(i)
		if strings.EqualFold(field.Name, tree[0]) {
			//log.Printf("deleteInterfaceStruct matching %s with %s len:%d", field.Name, tree[0], len(tree))
			if len(tree) == 1 {
				//log.Printf("final item in %v matching %s with %s", v.Elem().Field(i).Kind(), field.Name, tree[0])

			}

			switch v.Elem().Field(i).Kind() {
			case reflect.Map, reflect.Slice, reflect.Ptr:
				// we only create a new element of this type if they are zero
				if isEmpty(v.Elem().Field(i)) {
					modNew, err := createStruct(v.Elem().Field(i))
					if err != nil {
						return fmt.Errorf("deleteInterfaceStruct failed to create instance for empty struct: %s", err)
					}
					//log.Printf("createStruct: set %T to %+v", v.Elem().Field(i).Interface(), modNew)
					v.Elem().Field(i).Set(reflect.ValueOf(modNew))
				}

			}

			return deleteValue(v.Elem().Field(i), tree[1:])
		}
	}
	//return fmt.Errorf("deleteInterfaceStruct type '%s' has not been found in the resource '%T'", tree[0], v2.Interface())
	return nil
}

// deleteInterfaceMap gets the value of an interface based on tree of a Map
func deleteInterfaceMap(mod interface{}, tree []string) error {
	_, _, v2, _ := getReflection(mod)
	//log.Printf("deleteInterfaceMap mod:%T type:%+v tree:%v ", v2.Interface(), v2.Kind(), tree)

	// Loop through all field of the structure
	for _, i := range v2.MapKeys() {
		if strings.EqualFold(i.String(), tree[0]) {
			v2.SetMapIndex(i, reflect.Value{})
		}
	}
	// blindly ignore if value never existed, we meet the request we want
	return nil
}

// deleteInterfaceSlice gets the value of an interface based on tree of a Slice
func deleteInterfaceSlice(mod interface{}, tree []string) error {
	_, _, v2, _ := getReflection(mod)
	//log.Printf("deleteInterfaceSlice mod:%T type:%+v tree:%v", v2.Interface(), v2.Kind(), tree)

	// Loop through all field of the structure
	if len(tree) == 0 {
		tree = append(tree, "0")
	}
	treeInt, err := strconv.Atoi(tree[0])
	if err != nil {
		return fmt.Errorf("deleteInterfaceSlice failed to convert '%s' in to a number: %s", tree[0], err)
	}

	for i := 0; i < v2.Len(); i++ {
		if i == treeInt {
			switch v2.Index(i).Kind() {
			case reflect.Ptr: // reflect.Map, reflect.Slice,
				return deleteInterface(v2.Index(i).Interface(), tree[1:])
			}
			return deleteValue(v2.Index(i), tree[1:])
		}
	}
	return nil
	//return fmt.Errorf("deleteInterfaceSlice slice '%s' has not been found in the resource '%T'", tree[0], v2.Interface())
}
