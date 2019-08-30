package controllers

import (
	"net/http"

	"github.com/denisbakhtin/panteon/models"
	"github.com/gin-gonic/gin"
)

//HomeGet handles GET / route
func HomeGet(c *gin.Context) {
	db := models.GetDB()
	page := models.Page{}
	var products []models.Product

	db.First(&page, getSetting("home_id"))
	db.Where("published = true and category_id != ?", getSetting("our_works")).Order("recommended desc, id desc").Preload("Images").Limit(18).Find(&products)
	h := DefaultH(c)
	h["Page"] = page
	h["Products"] = products

	c.HTML(http.StatusOK, "home/show", h)
}
