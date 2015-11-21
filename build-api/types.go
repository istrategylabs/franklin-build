package main

type ApiResponse struct {
	Status   string        `json:"status" binding:"required"`
	Code     int           `json:"code" binding:"required"`
	Messages []interface{} `json:"messages"`
	Result   interface{}   `json:"result"`
}
