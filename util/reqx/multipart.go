package reqx

import (
	"net/http"
	"strings"
	
	"github.com/daarlabs/hirokit/constant/contentType"
	"github.com/daarlabs/hirokit/constant/header"
)

func IsMultipart(req *http.Request) bool {
	return strings.Contains(req.Header.Get(header.ContentType), contentType.MultipartForm)
}
