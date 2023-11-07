package annotation

var _i18n = map[string]string{}

func i(key string) string {
	ret, ok := _i18n[key]
	if !ok {
		return key
	}
	return ret
}
