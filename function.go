package bigModel

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func Json_encode(data interface{}) string {
	jsondata, er := json.Marshal(data)
	if er != nil {
		fmt.Println(er.Error())
		return ""
	}
	jsondata = bytes.Replace(jsondata, []byte("\\u0026"), []byte("&"), -1)
	jsondata = bytes.Replace(jsondata, []byte("\\u003c"), []byte("<"), -1)
	jsondata = bytes.Replace(jsondata, []byte("\\u003e"), []byte(">"), -1)
	return string(jsondata)
}
