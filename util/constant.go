package util

const (
	AsanaHost = "https://app.asana.com"

	ContentType = "application/json"

	HttpGetMethod  = "GET"
	HttpPutMethod  = "PUT"
	HttpPostMethod = "POST"

	AsanaDateFormat = "2006-01-02"
)

var (
	AsanaHeaders = make(map[string]string)
)
