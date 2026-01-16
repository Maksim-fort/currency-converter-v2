package model

// ConvertRequest - запрос на конвертацию
type ConvertRequest struct {
	From   string  `form:"from" binding:"required,len=3"` // form вместо json!
	To     string  `form:"to" binding:"required,len=3"`
	Amount float64 `form:"amount" binding:"required,min=0.01"`
}

// ConvertResponse - ответ на конвертацию
type ConvertResponse struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Rate   float64 `json:"rate"`
	Result float64 `json:"result"`
}

// ErrorResponse - структура для ошибок
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}
