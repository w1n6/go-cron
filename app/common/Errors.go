package common

import "errors"

var (
	Err_Lock_Already_Required = errors.New("锁已被占用") //锁已被占用

	Err_No_Local_IP_Found = errors.New("没有找到网卡IP") //没有找到网卡IP
)
