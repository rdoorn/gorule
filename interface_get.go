package gorule

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

// getInterface gets the value of an interface based on tree
func getInterface(mod interface{}, tree []string) (interface{}, error) {
	_, _, v2, t2 := getReflection(mod)
	log.Printf("getInterface mod:%T type:%+v tree:%v", v2.Interface(), v2.Kind(), tree)

	switch v2.Kind() {
	case reflect.String:
		return v2.String(), nil
	case reflect.Int:
		return v2.Int(), nil
	case reflect.Bool:
		return v2.Bool(), nil
	case reflect.Struct:
		return getInterfaceStruct(mod, tree)
	case reflect.Map:
		return getInterfaceMap(mod, tree)
	case reflect.Slice:
		return getInterfaceSlice(mod, tree)
	default:
		return "", fmt.Errorf("getInterface type '%s' has not been found in the resource '%T'", tree[0:], t2.String())
	}
}

// getInterfaceStruct gets the value of an interface based on tree of a Structure
func getInterfaceStruct(mod interface{}, tree []string) (interface{}, error) {
	v, _, v2, t2 := getReflection(mod)
	log.Printf("getInterfaceStruct mod:%T type:%+v tree:%v", v2.Interface(), v2.Kind(), tree)
	// Loop through all field of the structure
	for i := 0; i < t2.NumField(); i++ {
		field := t2.Field(i)
		if strings.EqualFold(field.Name, tree[0]) {
			return getInterface(v.Elem().Field(i).Interface(), tree[1:])
		}
	}
	return "", fmt.Errorf("getInterfaceStruct type '%s' has not been found in the resource '%T'", tree[0], t2.String())
}

// getInterfaceMap gets the value of an interface based on tree of a Map
func getInterfaceMap(mod interface{}, tree []string) (interface{}, error) {
	_, _, v2, t2 := getReflection(mod)
	log.Printf("getInterfaceMap mod:%T type:%+v tree:%v", v2.Interface(), v2.Kind(), tree)

	// Loop through all field of the structure
	for _, i := range v2.MapKeys() {
		log.Printf("map match %s, %s", v2.MapIndex(i).Interface(), tree[0])
		if strings.EqualFold(i.String(), tree[0]) {
			return getInterface(v2.MapIndex(i).Interface(), tree[1:])
		}
	}
	return "", fmt.Errorf("getInterfaceMap type '%s' has not been found in the resource '%T'", tree[0], t2.String())
}

// getInterfaceSlice gets the value of an interface based on tree of a Slice
func getInterfaceSlice(mod interface{}, tree []string) (interface{}, error) {
	_, _, v2, t2 := getReflection(mod)
	log.Printf("getInterfaceSlice mod:%T type:%+v tree:%v", v2.Interface(), v2.Kind(), tree)

	// Loop through all field of the structure
	if len(tree) == 0 {
		tree = append(tree, "0")
	}
	treeInt, err := strconv.Atoi(tree[0])
	if err != nil {
		return "", fmt.Errorf("getInterfaceSlice failed to convert '%s' in to a number: %s", tree[0], err)
	}

	for i := 0; i < v2.Len(); i++ {
		if i == treeInt {
			return getInterface(v2.Index(i).Interface(), tree[1:])
		}
	}
	return "", fmt.Errorf("getInterfaceSlice slice '%s' has not been found in the resource '%T'", tree[0], t2.String())
}
