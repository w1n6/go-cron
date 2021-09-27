package common

const (
	RequestSuccess = iota //请求成功 0
	RequestFail           //请求失败 1
)

type HttpResponse struct {
	HasErr int         `json:"haserr"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}
