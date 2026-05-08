package dto

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func Success() *Response {
	return &Response{
		Status: "success",
	}
}

func Error(msg string) *Response {
	return &Response{
		Status: "error",
		Error:  msg,
	}
}
