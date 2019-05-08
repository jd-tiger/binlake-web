package models

// RespData 返回总共的行数 以及具体行当中的数据
type RespData struct {
	Total int64       `json:"total"`
	Rows  interface{} `json:"rows"`
}

// RespMsg 返回code 以及错误信息
type RespMsg struct {
	Code    int64       `json:"code"`
	Message interface{} `json:"message"`
}

// RespErrMsg 返回code 以及string类型错误信息
type RespErrMsg struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

// RespData 返回code string类型错误信息 以及data数据
type RespMsgData struct {
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
