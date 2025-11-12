package utils

func Scheme(s bool) string {
	if s {
		return "https"
	} else {
		return "http"
	}
}
