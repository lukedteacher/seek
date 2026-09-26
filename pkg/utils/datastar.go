package utils

import "fmt"

func QueryValidate(url string) string {
	return fmt.Sprintf("@query('%s/validate')", url)
}

func DataBind(signal string) string {
	return fmt.Sprintf("data-bind:%s")
}
