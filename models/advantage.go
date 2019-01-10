package models

//Advantage type contains advantages info
type Advantage struct {
	Model

	Title       string `form:"title"`
	Description string `form:"description"`
}
