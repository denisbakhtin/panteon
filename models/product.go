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

//BeforeCreate gorm hook
func (p *Product) BeforeCreate() (err error) {
	if strings.TrimSpace(p.Slug) == "" {
		p.Slug = createSlug(fmt.Sprintf("%s %s", p.Title, p.Code))
	}
	return
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
