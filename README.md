# gorule

script based rules to modify interfaces. this will allow you to adjust your interfaces using scripts and variables

# example

blow is an example where a script could be used to change a http request. this can be used in proxies, where you allow custom rules to affect the http Request.

This allows you to let your customer modify your interfaces before or after transmission if you allow them to write the scripts using a config file.

```
package main

import (
  "log"
  "net/http"

  "github.com/rdoorn/gorule"
)

func main() {
  req, _ := http.NewRequest("GET", "http://www.tweakers.net", nil)
  script := ` request.url.path = "/about" `
  err := gorule.Parse(map[string]interface{}{"request": req}, []byte(script))
  if err != nil {
    log.Fatal(err)
  }

  log.Printf("request path modified: %s", req.URL.Path)
}
```
