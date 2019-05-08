package models

import (
	"encoding/json"
	"io/ioutil"
)

var (
	msgs = map[int]string{}
)

type errorItem struct {
	Code    int
	Message string
}

const (
	conf        = "conf/errors.json"
	UnkownError = "未知错误"
)

//LoadMessage 读取配置文件中的error message.
func LoadMessage() error {
	data, err := ioutil.ReadFile(conf)
	if err != nil {
		return err
	}

	var items []errorItem

	if err = json.Unmarshal(data, &items); err != nil {
		return err
	}

	for _, i := range items {
		msgs[i.Code] = i.Message
	}

	return nil
}

//Message 返回code对应的中文.
func Message(code int) string {
	if msg, ok := msgs[code]; ok {
		return msg
	}
	return UnkownError
}

/*
func main() {
	err := LoadMessage()
	if err != nil {
		panic(err.Error())
	}

    fmt.Printf("101:%v\n",  Message(101))
    fmt.Printf("999:%v\n",  Message(999))
}
*/
