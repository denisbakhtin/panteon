package models

//Setting type contains settings info
type Setting struct {
	Model

	Code  string `binding:"required" form:"code"`
	Title string `form:"title"`
	Value string `form:"value"`
}
