package models

//Image type contains image info
type Image struct {
	Model

	URL       string  `form:"url"`
	ProductID uint64  `form:"product_id"`
	Hash      string  `gorm:"-" form:"-"`
	Product   Product `gorm:"association_autoupdate:false;association_autocreate:false" binding:"-" form:"-"`
}
