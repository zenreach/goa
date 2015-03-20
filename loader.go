package goa

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

// LoadRequestBody decodes the request body. It returns the decoded content or an array of decoded
// contents in the case of a multipart body.
// The following content types are  supported:
// application/json, text/json, <anything>+json: body is decoded with the JSON decoder.
// application/x-www-form-urlencoded: body is read as a url encoded form.
// multipart/<anything>: each part is decoded using the decoder returned by applying this same
// algorithm to the part content-type header.
// Returns an error if the content type is not supported or decoding fails.
func LoadRequestBody(r *http.Request) (interface{}, error) {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("invalid request media type: %s", err)
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(r.Body, params["boundary"])
		var contents []interface{}
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				return contents, nil
			}
			if err != nil {
				return nil, fmt.Errorf("fail to read part enveloppe: %s", err)
			}
			c, err := loadSingleBody(p.Header.Get("Content-Type"), p)
			if err != nil {
				return nil, fmt.Errorf("fail to decode part body: %s", err)
			}
			contents = append(contents, c)
		}
	}
	return loadSingleBody(mediaType, r.Body)
}

// loadSingleBody is a helper function used by LoadRequestBody to decode the content of a single
// HTTP request body encoded using the media type identified by mt. See LoadRequestBody for more
// details.
func loadSingleBody(mt string, body io.Reader) (interface{}, error) {
	if strings.Contains(mt, "form-urlencoded") {
		maxFormSize := int64(1<<63 - 1)
		b, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("fail to read form body: %s", err)
		}
		if int64(len(b)) > maxFormSize {
			return nil, fmt.Errorf("request body too large")
		}
		vs, err := url.ParseQuery(string(b))
		if err != nil {
			return nil, fmt.Errorf("fail to decode form body: %s", err)
		}
		values := make(map[string]interface{})
		for n, v := range vs {
			values[n] = v
		}
		return values, nil
	} else if strings.HasSuffix(mt, "json") {
		decoder := json.NewDecoder(body)
		var decoded interface{}
		err := decoder.Decode(&decoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %s", err)
		}
		return decoded, nil
	} else {
		return nil, fmt.Errorf("unsupported content type '%s'", mt)
	}
}
