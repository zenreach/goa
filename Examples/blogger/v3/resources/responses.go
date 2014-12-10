package v3

// Common responses

var responses = struct {
	badRequest           Response
	unauthorizedResponse Response
}{
	badRequest: Response{
		Status:    400,
		MediaType: errorMediaType,
		Headers:   requiredHeaders,
	},
	unauthorizedResponse: Response{
		Status:    401,
		MediaType: errorMediaType,
		Headers:   requiredHeaders,
	},
}
