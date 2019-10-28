package gorule

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

// modifyInterface changes the interface, based on the tree input to the set value
func modifyInterface(mod interface{}, tree []string, value string) error {

	log.Printf("modifyInterface mod:%T tree:%v value:%s", mod, tree, value)
	_, _, v2, _ := getReflection(mod)

	//log.Printf("Modifying %T ", mod)
	switch v2.Kind() {
	case reflect.String:
	case reflect.Int:
	default:
		log.Printf("Kind: %v %T", v2.Kind(), v2.Kind())
		return modifyInterfaceStruct(mod, tree, value)

	}
	// at this point we expect a struct of some kind
	// loop through the fields of the struct
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

func modifyInterfaceStruct(mod interface{}, tree []string, value string) error {
	log.Printf("modifyInterfaceStruct mod:%T tree:%v value:%s", mod, tree, value)
	t := reflect.TypeOf(mod)
	v := reflect.ValueOf(mod)

	var v2 reflect.Value
	var t2 reflect.Type

	// convert pointer to non-pointer
	if v.Kind() == reflect.Ptr {
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

			log.Printf("KIND: %+v", v.Elem().Field(i).Kind())

			switch v.Elem().Field(i).Kind() {
			case reflect.Bool:
				if strings.EqualFold(value, "true") {
					v.Elem().Field(i).SetBool(true)
				} else {
					v.Elem().Field(i).SetBool(false)
				}
			case reflect.String:
				v.Elem().Field(i).SetString(value)
			case reflect.Int:
				newValue, err := strconv.Atoi(value)
				if err != nil {
					return err
				}
				v.Elem().Field(i).SetInt(int64(newValue))
			case reflect.Map:
				log.Printf("modifyInterfaceStruct MAP")

				// we only create MAP if they are zero
				log.Printf("v2: %T", v.Elem().Field(i).Interface())
				log.Printf("isEmpty: %t", isEmpty(v.Elem().Field(i)))
				if isEmpty(v.Elem().Field(i)) {
					modNew, err := createStruct(v.Elem().Field(i))
					if err != nil {
						return fmt.Errorf("modifyInterface failed to create instance for empty struct: %s", err)
					}
					log.Printf("modNew: %T", modNew)
					log.Printf("IsSettable %+v", v.Elem().Field(i).CanSet())
					v.Elem().Field(i).Set(reflect.ValueOf(modNew))

				}

				/* WORKS
				m := v.Elem().Field(i).Interface()
				// test if the value is empty, if so try to fill it
				if reflect.DeepEqual(m, reflect.Zero(reflect.TypeOf(m)).Interface()) {

					switch reflect.TypeOf(v.Elem().Field(i).Interface()).String() {
					case "http.Header":
						z := make(http.Header)
						v.Elem().Field(i).Set(reflect.ValueOf(z))
					default:
						return fmt.Errorf("cannot create field of type: %s", v.Elem().Field(i).Kind().String())
					}

				}
				*/
				log.Printf("modifyInterfaceStruct Move in to Map")
				err := modifyInterfaceMap(v.Elem().Field(i).Interface(), tree[1:], value)
				return err
			case reflect.Ptr:
				log.Printf("modifyInterfaceStruct Move in to Ptr")

				// we only create pointers if they are zero
				log.Printf("v2: %T", v.Elem().Field(i).Interface())
				log.Printf("isEmpty: %t", isEmpty(v.Elem().Field(i)))
				if isEmpty(v.Elem().Field(i)) {
					modNew, err := createStruct(v.Elem().Field(i))
					if err != nil {
						return fmt.Errorf("modifyInterface failed to create instance for empty struct: %s", err)
					}
					log.Printf("modNew: %T", modNew)
					log.Printf("IsSettable %+v", v.Elem().Field(i).CanSet())
					v.Elem().Field(i).Set(reflect.ValueOf(modNew))

				}

				/*v2 = reflect.Indirect(v)
				log.Printf("PTR testing if kind is empty")
				log.Printf("PTR v2: %T", v2.Elem().Field(i).Interface())
				log.Printf("isEmpty: %t", isEmpty(v2.Elem().Field(i)))
				if isEmpty(v2) {
					modNew, err := createStruct(v2.Elem().Field(i))
					if err != nil {
						return fmt.Errorf("modifyInterface failed to create instance for empty struct: %s", err)
					}
					log.Printf("modNew: %T", modNew)
					log.Printf("IsSettable %+v", v2.Elem().Field(i).CanSet())
					v.Elem().Field(i).Set(reflect.ValueOf(modNew))

				}*/

				return modifyInterface(v.Elem().Field(i).Interface(), tree[1:], value)
			case reflect.Struct:
				log.Printf("modifyInterfaceStruct Move in to Struct")

				return modifyInterface(v.Elem().Field(i).Interface(), tree[1:], value)
			default:
				return fmt.Errorf("modifyInterfaceStruct type '%s' has not been implemented yet in the interface modifier", v.Elem().Field(i).Kind())
			}

		}
	}
	return nil
}

func modifyInterfaceMap(mod interface{}, tree []string, value string) error {

	log.Printf("Is Zero?, %t", reflect.DeepEqual(mod, reflect.Zero(reflect.TypeOf(mod)).Interface()))
	t := reflect.TypeOf(mod)
	v := reflect.ValueOf(mod)

	log.Printf("Map keys: %+v searching for: %s setting: %s", v.MapKeys(), tree[1:], value)

	//log.Printf("MAP t: %+v", t)
	//log.Printf("MAP v: %+v", v)

	log.Printf("map key check 0")
	for _, e := range v.MapKeys() {
		log.Printf("map key check 1")
		switch e.Interface().(type) {
		case string:
			log.Printf("map key check 2")
			if strings.EqualFold(e.String(), tree[0]) {
				log.Printf("modifying key: %s", e)
				m := v.MapIndex(e)
				switch m.Interface().(type) {
				case int:
					fmt.Println(e, t)
				case []string:
					// headers have an [][]string setup in go, generally they are all [0], so lets assume for now
					//log.Printf("map field before: %+v", v.MapIndex(e).Interface())
					err := modifyInterfaceSlice(m.Interface(), tree[1:], value)
					//log.Printf("map field after: %+v", v.MapIndex(e).Interface())
					//log.Printf("map field keys after: %+v", v.MapKeys())
					return err

					//v.Elem().Field(i).MapIndex(e).Field(0).SetString(value)
					//fmt.Println(e, t)
				case string:
					fmt.Println(e, t)
				case bool:
					fmt.Println(e, t)
				default:
					log.Printf("not found: %+v %T", m, m)

				}
			}
		default:
			return fmt.Errorf("modifyInterfaceMap of type: %T is not yet supported", e.Interface())
		}
	}

	log.Printf("map key check 4")
	// we have not adjusted anything yet, so create a new entry

	// get the kind of the map key
	log.Printf("map key kind is: %s", t.Key().Kind())
	switch t.Key().Kind() {
	case reflect.String:
		// get the kind of the map value
		log.Printf("map value kind is: %s", t.Elem())
		switch t.Elem().Kind() {
		case reflect.Slice:
			log.Printf("map int is: %+v, %+v LEN:%d", t, t, v.Len())
			if v.Len() == 0 {
				switch t.String() {
				case "http.Header":
					/*
						v := reflect.ValueOf(z)
						t = reflect.TypeOf(z)
					*/

					/*z := make(map[string][]string)
					mod = unsafe.Pointer(&z)
					v = reflect.ValueOf(mod)
					t = reflect.TypeOf(mod)*/

					//myType := reflect.TypeOf(z)
					//slice := reflect.MakeSlice(reflect.SliceOf(myType), 1, 1)
					//log.Printf("slice is now: %s %T", slice, slice)
					//log.Printf("slice elem is now: %s %T", t.Elem(), t.Elem())
					// Create a pointer to a slice value and set it to the slice
					//t.Elem().Set(slice)
					//}
					//z := make(http.Header)
					//log.Printf("T is now: %s %T", t, t)
					//z := reflect.Indirect(reflect.New(t))
					//log.Printf("Z is now: %+v %T", z, z)
					//t2 := reflect.TypeOf(z)
					//v = reflect.Indirect(reflect.New(t))
					//v2 := reflect.ValueOf(z)
					//v.Set(v2)
					//v2 := reflect.ValueOf(z)
					//log.Printf("map2 key kind is: %s", t2.Key().Kind())
					//log.Printf("map2 value kind is: %s", t2.Elem())
					//log.Printf("map2 int is: %+v, %+v", t2, t2)
					//v2 := reflect.ValueOf(z)
					//obj := reflect.Indirect(v)
					//obj.Set(v)

					//log.Printf("can addr %t", v.CanAddr())
					//log.Printf("can set %t", v.CanSet())

					//log.Printf("ptr of mod: %+v", &mod)
					//log.Printf("ptr of z: %+v", interface{ z })
					//mod = z
				}
			}

			// adding key to exising map
			log.Printf("adding key to existing map of len:%d", v.Len())

			e := reflect.Indirect(reflect.New(t.Elem())).Interface()
			switch e.(type) {
			case []string:
				e = append(e.([]string), value)
			default:
				log.Printf("Unknown type %T", e)
				return fmt.Errorf("unknown type...")
			}
			v.SetMapIndex(reflect.ValueOf(tree[0]), reflect.ValueOf(e))
			return nil

		}
	}
	return nil
}

func modifyInterfaceSlice(mod interface{}, tree []string, value string) error {
	//t := reflect.TypeOf(mod)
	v := reflect.ValueOf(mod)

	log.Printf("Slice keys: %+v searching for: %s setting: %s", v.Len(), tree, value)
	if len(tree) == 0 {
		tree = append(tree, "0")
	}

	index, err := strconv.Atoi(tree[0])
	if err != nil {
		return fmt.Errorf("expected index number for slice but got:%s, error:%s", tree, err)
	}

	if index > v.Len() {
		return fmt.Errorf("index otu of range, your setting %d where the max is %d", index, v.Len())
	}

	//log.Printf("Slice index kind: %+v", v.Index(index).Kind())
	//log.Printf("Slice data before: %+v", v)

	switch v.Index(index).Kind() {
	case reflect.String:
		v.Index(index).SetString(value)

	}
	//log.Printf("Slice data after: %+v", v)

	//log.Printf("Len t: %+v", t)
	//log.Printf("Len v: %+v", v)

	//for _, e := range v.NumField() {
	//for i := 0; i < v.Len(); i++ {
	//if strings.EqualFold(e, tree[1:]) {
	/*v := v.MapIndex(e)
	switch v.Interface().(type) {
	case int:
		fmt.Println(e, t)
	case []string:
		// headers have an [][]string setup in go, generally they are all [0], so lets assume for now
		log.Printf("map field: %+v", v.Interface().([]string))
		return modifyInterfaceMap(v.Interface(), tree[1:], value)
		//v.Elem().Field(i).MapIndex(e).Field(0).SetString(value)
		//fmt.Println(e, t)
	case string:
		fmt.Println(e, t)
	case bool:
		fmt.Println(e, t)
	default:
		log.Printf("not found: %+v %T", v, v)

	}*/
	//}
	//}
	return nil
}
