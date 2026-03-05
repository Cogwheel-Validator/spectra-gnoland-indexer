package humatypes

type ConvertFromBase64toBase64UrlInput struct {
	TxHash64 string `query:"tx_hash_64" doc:"Input to convert" required:"true" minLength:"44" maxLength:"44"`
}

type ConvertFromBase64UrlToBase64Input struct {
	TxHash64Url string `query:"tx_hash_64_url" doc:"Input to convert" required:"true" minLength:"44" maxLength:"44"`
}

type ConvertFromBase64toBase64UrlBody struct {
	TxHash64Url string `json:"tx_hash_64_url" doc:"Output of base64 to base64url conversion" required:"true"`
}
type ConvertFromBase64toBase64UrlOutput struct {
	Body ConvertFromBase64toBase64UrlBody
}

type ConvertFromBase64UrlToBase64Body struct {
	TxHash64 string `json:"tx_hash_64" doc:"Output of base64url to base64 conversion" required:"true"`
}
type ConvertFromBase64UrlToBase64Output struct {
	Body ConvertFromBase64UrlToBase64Body
}
