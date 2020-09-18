package request

import (
	"github.com/restic/restic/internal/restic"
	"net/http"
	"strings"
)

func GetParams(r *http.Request) map[string]string {
	params := map[string]string{}

	for key, vals := range r.URL.Query() {
		val := ""
		if len(vals) > 0 {
			val = vals[0]
		}
		params[key] = val
	}

	return params
}

func GetCommaSepParams(key string, r *http.Request) (params []string) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return
	}

	return strings.Split(value, ",")
}

func GetTags(key string, r *http.Request) (tags restic.TagLists) {
	tagsFlat := GetCommaSepParams(key, r)
	for _, tagFlag := range tagsFlat {
		tags = append(tags, []string{tagFlag})
	}

	return
}

func GetBoolParam(key string, r *http.Request) bool {
	value := r.URL.Query().Get(key)
	return value != ""
}
