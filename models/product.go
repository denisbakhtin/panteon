package models

import (
	"fmt"
	"strings"
)

//Product type contains product info
type Product struct {
	Model

	Title           string   `form:"title"`
	Description     string   `form:"description"`
	MetaKeywords    string   `form:"meta_keywords"`
	MetaDescription string   `form:"meta_description"`
	CategoryID      uint64   `form:"category_id"`
	Published       bool     `form:"published"`
	Slug            string   `form:"slug"`
	Code            string   `form:"code"`
	Category        Category `gorm:"save_associations:false" binding:"-" form:"-"`
	ImageIds        []uint64 `form:"imageids" gorm:"-"` //hack
	DefaultImageID  uint64   `form:"defaultimageid"`
	Recommended     bool     `form:"recommended"`
	Images          []Image
	Orders          []Order `gorm:"save_associations:false" binding:"-" form:"-"`
}

//URL returns article url
func (p *Product) URL() string {
	return fmt.Sprintf("/p/%d-%s", p.ID, p.Slug)
}

//BeforeSave gorm hook
func (p *Product) BeforeSave() (err error) {
	if strings.TrimSpace(p.Slug) == "" {
		p.Slug = createSlug(fmt.Sprintf("%s %s", p.Title, p.Code))
	}
	return
}

//DefaultImage returns url of the default product img
func (p *Product) DefaultImage() string {
	img := Image{}
	db.First(&img, p.DefaultImageID)
	return img.URL
}

//DefaultImagePreview returns url of the default product img thumbnail
func (p *Product) DefaultImagePreview() string {
	img := Image{}
	db.First(&img, p.DefaultImageID)
	if len(img.PreviewURL) > 0 {
		return img.PreviewURL
	}
	return img.URL
}

//Breadcrumbs returns a list of product breadcrumbs
func (p *Product) Breadcrumbs() []Breadcrumb {
	var par, gpar, ggpar, gggpar Category
	db.First(&par, p.CategoryID)
	gpar = par.GetParent()
	ggpar = gpar.GetParent()
	gggpar = ggpar.GetParent()
	res := make([]Breadcrumb, 0, 10)
	res = append(res, Breadcrumb{Title: "Главная", URL: "/"})
	if gggpar.ID != 0 {
		res = append(res, Breadcrumb{Title: gggpar.Title, URL: gggpar.URL()})
	}
	if ggpar.ID != 0 {
		res = append(res, Breadcrumb{Title: ggpar.Title, URL: ggpar.URL()})
	}
	if gpar.ID != 0 {
		res = append(res, Breadcrumb{Title: gpar.Title, URL: gpar.URL()})
	}
	if par.ID != 0 {
		res = append(res, Breadcrumb{Title: par.Title, URL: par.URL()})
	}
	res = append(res, Breadcrumb{Title: p.Title})
	return res
}
