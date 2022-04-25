package utils

import "encoding/json"

func PrettyLog(res interface{}) string {
	s, _ := json.MarshalIndent(res, "", "\t")
	return string(s)
}
