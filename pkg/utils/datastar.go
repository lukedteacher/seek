package utils

import "fmt"

func Query(url string) string {
	return fmt.Sprintf("@query('%s')", url)
}

func QueryValidate(url string) string {
	return Query(url + "/validate")
}

func DataBind(signal string) string {
	return fmt.Sprintf("data-bind:%s", signal)
}

func DataSignals(signal string) string {
	return fmt.Sprintf("data-signals:%s", signal)
}

func OnInput(action string) string {
	return fmt.Sprintf("data-on:input__debounce.250ms=%s", action)
}

func OnInputQuery(url string) string {
	return OnInput(Query(url))
}
