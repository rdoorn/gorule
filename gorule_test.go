package gorule

import (
	"fmt"
	"log"
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

	// set tls client certificate
	scriptTest{
		interfaces: map[string]interface{}{},
		script: []byte(`
					var testvalue 1
					if $(testvalue) == 1 {
						//var resultvalue 1
						log "got 1"
					} elseif $(testvalue) == 2 {
						//var resultvalue 2
						log "got 2"
					} else {
						//var resultvalue 3
						log "got 3"
					}
				`),
		result: map[string]interface{}{
			"resultValue": 1,
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
	log.Printf("....... new test .......")
	err := parse(i, script)
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

/*
var scripts = map[string][]byte{
	"log": []byte(
		`
			var tmp 200
			response.StatusCode = 333
			request.url.host = localhost
			request.proto = "HTTP/3.0"
			response.header.cache-control = "none; bla"
			response.header.x-custom = "script power"
      if $(response.StatusCode) == $(tmp) {
        log "hello world"
			}
			//request.url.path = "/bla/bla2/path"
      } elseif 2 == 3 {
        log "hello elsif world"
      } else {
        log "hello else world"
			}

    `,
	),
	"modify_method": []byte(
		`
      if method == get {
        method = put
      }
    `,
	),
}

func TestRules(t *testing.T) {
	request, err := http.NewRequest("GET", "http://www.ghostbox.org", nil)
	client := &http.Client{}
	response, err := client.Do(request)
	//response := &http.Response{
	//Header: make(http.Header),
	//
	//response.Header.Add("key", "value")
	log.Printf("start request: %+v err:%s", request, err)

	err = parse(
		map[string]interface{}{
			"request":  request,
			"response": response,
		},
		scripts["log"],
	)
	assert.Nil(t, err)

	log.Printf("end request: %+v err:%s", request, err)
	log.Printf("end response: %+v err:%s", response, err)

}

func TestWord(t *testing.T) {
	script := []byte("abc def ghi jkl")

	parser := NewParser(script)
	res, err := parser.Word()
	assert.Nil(t, err)
	assert.Equal(t, "abc", res)
	res, err = parser.Word()
	assert.Nil(t, err)
	assert.Equal(t, "def", res)
	res, err = parser.Word()
	assert.Nil(t, err)
	assert.Equal(t, "ghi", res)
	res, err = parser.Word()
	assert.Nil(t, err)
	assert.Equal(t, "jkl", res)
}

*/
