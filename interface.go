package gorule

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
)

func isEmpty(t reflect.Value) bool {
	m := t.Interface()
	if reflect.DeepEqual(m, reflect.Zero(reflect.TypeOf(m)).Interface()) {
		log.Printf("value is zeo")
		return true
	}
	return false
}

func createStruct(t reflect.Value) (interface{}, error) {
	switch reflect.TypeOf(t.Interface()).String() {
	case "http.Header":
		z := make(http.Header)
		return z, nil
	case "*url.URL":
		z := &url.URL{}
		return z, nil
	case "*tls.ConnectionState":
		z := &tls.ConnectionState{}
		return z, nil
	case "[]*x509.Certificate":
		//var v []*certificate
		//z := &x509.Certificate{}
		//z := make([]*x509.Certificate, 1)
		//z = append(z, &x509.Certificate{})
		z := []*x509.Certificate{
			&x509.Certificate{Signature: []byte("aa:bb:cc")},
		}
		return z, nil
	case "[]uint8":
		z := []uint8{}
		return z, nil
	default:
		return nil, fmt.Errorf("cannot create field of type: %T", t.Interface())
	}
}

// getReflection returns the reflection interfaces, both with and without pointers
func getReflection(i interface{}) (reflect.Value, reflect.Type, reflect.Value, reflect.Type) {
	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

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
	return v, t, v2, t2
}
