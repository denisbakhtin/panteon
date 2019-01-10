package models

//Order type contains buy order info
type Order struct {
	Model

	FirstName  string    `form:"first_name"`
	MiddleName string    `form:"middle_name"`
	LastName   string    `form:"last_name"`
	Email      string    `form:"email"`
	Phone      string    `form:"phone"`
	Comment    string    `form:"comment"`
	Products   []Product `gorm:"many2many:order_products;association_autoupdate:false;association_autocreate:false" binding:"-" form:"-"`
}
