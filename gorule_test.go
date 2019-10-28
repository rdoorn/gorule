package gorule

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type scriptTest struct {
	interfaces map[string]interface{}
	script     []byte
	result     map[string]interface{}
}

var scriptTests = []scriptTest{
	// set string
	scriptTest{
		interfaces: map[string]interface{}{
			"request": &http.Request{},
		},
		script: []byte(`
			request.proto = "HTTP/1.9"
		`),
		result: map[string]interface{}{
			"request.proto": "HTTP/1.9",
		},
	},
	// set bool
	scriptTest{
		interfaces: map[string]interface{}{
			"request": &http.Request{},
		},
		script: []byte(`
				request.close = true
			`),
		result: map[string]interface{}{
			"request.close": true, // false by default
		},
	},

	// set url.path
	scriptTest{
		interfaces: map[string]interface{}{
			"request": &http.Request{},
		},
		script: []byte(`
					request.url.path = "/status"
				`),
		result: map[string]interface{}{
			"request.url.path": "/status",
		},
	},

	// set custom Header
	scriptTest{
		interfaces: map[string]interface{}{
			"request": &http.Request{},
		},
		script: []byte(`
				request.header.x-custom = "hello world"
			`),
		result: map[string]interface{}{
			"request.header.x-custom": "hello world",
		},
	},

	// set int64 (content length)
	scriptTest{
		interfaces: map[string]interface{}{
			"request": &http.Request{},
		},
		script: []byte(`
					request.contentlength = 10
				`),
		result: map[string]interface{}{
			"request.contentlength": int64(10), // default is 0
		},
	},

	// set int (statuscode
	scriptTest{
		interfaces: map[string]interface{}{
			"response": &http.Response{},
		},
		script: []byte(`
					response.statuscode = 456
				`),
		result: map[string]interface{}{
			"response.statuscode": int(456), // default is 0
		},
	},

	// set tls client certificate
	scriptTest{
		interfaces: map[string]interface{}{
			"request": &http.Request{},
		},
		script: []byte(`
					request.tls.peercertificates.0.Signature = "11:22:33:44:55:66:77:88"
				`),
		result: map[string]interface{}{
			"request.tls.peercertificates.0.Signature": "11:22:33:44:55:66:77:88",
		},
	},

	// match if
	scriptTest{
		interfaces: map[string]interface{}{},
		script: []byte(`
					var testvalue 1
					if $(testvalue) == 1 {
						testvalue = 10
					} elseif $(testvalue) == 2 {
						testvalue = 20
					} else {
						testvalue = 30
					}
				`),
		result: map[string]interface{}{
			"testvalue": "10",
		},
	},

	// match ifelse
	scriptTest{
		interfaces: map[string]interface{}{},
		script: []byte(`
					var testvalue 2
					if $(testvalue) == 1 {
						testvalue = 10
					} elseif $(testvalue) == 2 {
						testvalue = 20
					} else {
						testvalue = 30
					}
				`),
		result: map[string]interface{}{
			"testvalue": "20",
		},
	},

	// match else
	scriptTest{
		interfaces: map[string]interface{}{},
		script: []byte(`
					var testvalue 3
					if $(testvalue) == 1 {
						testvalue = 10
					} elseif $(testvalue) == 2 {
						testvalue = 20
					} else {
						testvalue = 30
					}
				`),
		result: map[string]interface{}{
			"testvalue": "30",
		},
	},

	// ignore comments
	scriptTest{
		interfaces: map[string]interface{}{},
		script: []byte(`
					var testvalue 3
					// testvalue = 4
					# testvalue = 5
					/*
						testvalue = 6
						*/
				`),
		result: map[string]interface{}{
			"testvalue": "3",
		},
	},

	// regex match
	scriptTest{
		interfaces: map[string]interface{}{
			"request": &http.Request{
				Header: map[string][]string{
					"Referer": []string{"http://example.com/"},
				},
			},
		},
		script: []byte(`
							var testvalue 1
							if $(request.header.referer) regex "e[x]+..ple" {
								testvalue = 3
							}
				`),
		result: map[string]interface{}{
			"testvalue": "3",
		},
	},
}

func TestScript(t *testing.T) {
	// execute tests
	t.Run("scriptTests", func(t *testing.T) {
		for id, script := range scriptTests {
			t.Run(fmt.Sprintf("scriptTests/%d", id), func(t *testing.T) {
				runTestScript(t, script.interfaces, script.script, script.result)
			})
		}
	})
}

func runTestScript(t *testing.T, i map[string]interface{}, script []byte, result map[string]interface{}) {
	//log.Printf("....... new test .......")
	err := Parse(i, script)
	assert.Nil(t, err, fmt.Sprintf("script:%s execution returned error", script))
	if err == nil {
		for testVariable, expected := range result {
			testTree := strings.Split(testVariable, ".")
			returned, err := getInterface(i[testTree[0]], testTree[1:])
			assert.Nil(t, err, fmt.Sprintf("script:%s getInterface of:%s returned error", script, testVariable))
			assert.Equal(t, expected, returned, fmt.Sprintf("script:%s get result of:%s returned incorrect result, got: %+v", script, expected, returned))
		}
	}
}
