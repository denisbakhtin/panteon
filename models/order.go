package models

//Order type contains product order info
type Order struct {
	Model

	Name      string  `form:"name"`
	Email     string  `form:"email"`
	Phone     string  `form:"phone"`
	Comment   string  `form:"comment"`
	ProductID uint64  `form:"product_id"`
	BackURL   string  `form:"back_url"`
	Product   Product `form:"-" binding:"-"`
}
