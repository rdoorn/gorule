package gorule

import (
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var scripts = map[string][]byte{
	"log": []byte(
		`
			var tmp 200
			/*response.StatusCode = 333
			request.url.host = localhost
			request.proto = "HTTP/3.0"
			response.header.cache-control = "none; bla"
			response.header.x-custom = "script power"*/
      if $(response.StatusCode) == $(tmp) {
        log "hello world"
			}
			//request.url.path = "/bla/bla2/path"
    /*  } elseif 2 == 3 {
        log "hello elsif world"
      } else {
        log "hello else world"
			}*/

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
	/*response := &http.Response{
		Header: make(http.Header),
	}*/
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
