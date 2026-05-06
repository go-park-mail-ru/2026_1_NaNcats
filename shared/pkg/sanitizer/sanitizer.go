package sanitizer

import "github.com/microcosm-cc/bluemonday"

var Policy *bluemonday.Policy

func init() {
	Policy = bluemonday.StrictPolicy()
}

type Sanitizable interface {
	Sanitize(p *bluemonday.Policy)
}
