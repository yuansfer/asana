package text

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

const (
	DataMention = `<a data-asana-gid="%s"/>`
)

//	`<body>
//		testing
//		<a data-asana-gid="1199521781039350"/>
//		<a data-asana-gid="1200773217477032"/>
//	</body>`
func Mention(gid string) string {
	return fmt.Sprintf(DataMention, gid)
}

func DeleteHtmlTags(src string) string {
	var (
		line   int
		v, des string
	)
	if "" == src || 0 == len(src) {
		return ""
	}
	reg := regexp.MustCompile(`<.+?>`)
	str := reg.ReplaceAllString(src, " ")
	s := strings.Split(str, "\n")
	for _, v = range s {
		if 0 == len(strings.TrimSpace(v)) {
			continue
		}
		des = des + fmt.Sprintf("\n%s", v)
		line++
	}
	if 0 == line {
		return ""
	}

	return html.UnescapeString(des)[1:]
}
