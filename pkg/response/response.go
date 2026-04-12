package response

//go:generate easyjson $GOFILE

//easyjson:json
type MessageResponse struct {
	Message string `json:"message"`
}
