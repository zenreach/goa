package status

import (
	"net/http"
	"strings"

	"github.com/raphael/goa"
)

// goa.Response factory methods
// Example usage:
//     r := status.Created().WithLocation(href)

func Continue() *goa.Response                   { return vanillaResponse(100) }
func Ok() *goa.Response                         { return vanillaResponse(200) }
func Created() *goa.Response                    { return vanillaResponse(201) }
func Accepted() *goa.Response                   { return vanillaResponse(202) }
func NonAuthoritative() *goa.Response           { return vanillaResponse(203) }
func NoContent() *goa.Response                  { return vanillaResponse(204) }
func ResetContent() *goa.Response               { return vanillaResponse(205) }
func PartialContent() *goa.Response             { return vanillaResponse(206) }
func MultipleChoices() *goa.Response            { return vanillaResponse(300) }
func MovedPermanently() *goa.Response           { return vanillaResponse(301) }
func Found() *goa.Response                      { return vanillaResponse(302) }
func SeeOther() *goa.Response                   { return vanillaResponse(303) }
func NotModified() *goa.Response                { return vanillaResponse(304) }
func UseProxy() *goa.Response                   { return vanillaResponse(305) }
func TemporaryRedirect() *goa.Response          { return vanillaResponse(307) }
func BadRequest() *goa.Response                 { return vanillaResponse(400) }
func Unauthorized() *goa.Response               { return vanillaResponse(401) }
func PaymentRequired() *goa.Response            { return vanillaResponse(402) }
func Forbidden() *goa.Response                  { return vanillaResponse(403) }
func NotFound() *goa.Response                   { return vanillaResponse(404) }
func MethodNotAllowed() *goa.Response           { return vanillaResponse(405) }
func NotAcceptable() *goa.Response              { return vanillaResponse(406) }
func ProxyAuthRequired() *goa.Response          { return vanillaResponse(407) }
func RequestTimeout() *goa.Response             { return vanillaResponse(408) }
func Conflict() *goa.Response                   { return vanillaResponse(409) }
func Gone() *goa.Response                       { return vanillaResponse(410) }
func LengthRequired() *goa.Response             { return vanillaResponse(411) }
func PreconditionFailed() *goa.Response         { return vanillaResponse(412) }
func RequestEntityTooLarge() *goa.Response      { return vanillaResponse(413) }
func RequestUriTooLong() *goa.Response          { return vanillaResponse(414) }
func UnsupportedMediaType() *goa.Response       { return vanillaResponse(415) }
func RequestRangeNotSatisfiable() *goa.Response { return vanillaResponse(416) }
func ExpectationFailed() *goa.Response          { return vanillaResponse(417) }
func InternalError() *goa.Response              { return vanillaResponse(500) }
func NotImplemented() *goa.Response             { return vanillaResponse(501) }
func BadGateway() *goa.Response                 { return vanillaResponse(502) }
func ServiceUnavailable() *goa.Response         { return vanillaResponse(503) }
func GatewayTimeout() *goa.Response             { return vanillaResponse(504) }
func HTTPVersionNotSupported() *goa.Response    { return vanillaResponse(505) }

// vanillaResponse returns a default response for the given HTTP status code
func vanillaResponse(status int) *goa.Response {
	return &goa.Response{Status: status, Body: strings.NewReader(http.StatusText(status))}
}
