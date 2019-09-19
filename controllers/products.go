package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/denisbakhtin/panteon/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

//ProductGet handles GET /p/:slug route
func ProductGet(c *gin.Context) {
	db := models.GetDB()
	product := models.Product{}

	idslug := c.Param("idslug")
	id := atouint(strings.Split(idslug, "-")[0])
	db.Preload("Images").First(&product, id)
	if product.ID == 0 || !product.Published {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	//redirect to canonical url
	if c.Request.URL.Path != product.URL() {
		c.Redirect(303, product.URL())
		return
	}

	h := DefaultH(c)
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	h["Title"] = product.Title
	h["Product"] = &product
	h["MetaKeywords"] = product.MetaKeywords
	h["MetaDescription"] = product.MetaDescription
	session.Save()
	c.HTML(http.StatusOK, "products/show", h)
}

//ProductIndex handles GET /admin/products route
func ProductIndex(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	query := strings.ToLower(c.Query("query"))
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	db := models.GetDB()
	var products []models.Product
	dbq := db.Model(&models.Product{})
	if len(query) > 0 {
		dbq = dbq.Where("lower(title) like ?", fmt.Sprintf("%%%s%%", query))
	}
	if limit == 0 {
		limit = 100
	}
	paginator := buildPaginator(dbq, c.Request.URL.Path, c.Request.URL.RawQuery, limit, page)
	dbq.Offset((page - 1) * limit).Limit(limit).
		Order("id asc").Preload("Category").Find(&products)
	h := DefaultH(c)
	h["Title"] = "Список продукции"
	h["Products"] = products
	h["Limit"] = limit
	h["Query"] = query
	h["Page"] = page
	h["Paginator"] = paginator
	c.HTML(http.StatusOK, "products/index", h)
}

//ProductNew handles GET /admin/new_product route
func ProductNew(c *gin.Context) {
	h := DefaultH(c)
	h["Title"] = "Новое изделие"
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	h["Product"] = models.Product{Published: true}
	session.Save()

	c.HTML(http.StatusOK, "products/form", h)
}

//ProductCreate handles POST /admin/new_product route
func ProductCreate(c *gin.Context) {
	db := models.GetDB()
	product := models.Product{}
	if err := c.ShouldBind(&product); err != nil {
		session := sessions.Default(c)
		session.AddFlash(err.Error())
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/new_product")
		return
	}
	db.Where("id in (?)", product.ImageIds).Find(&product.Images)
	if err := db.Create(&product).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/products")
}

//ProductEdit handles GET /admin/products/:id/edit route
func ProductEdit(c *gin.Context) {
	db := models.GetDB()
	product := models.Product{}
	db.Preload("Images").First(&product, c.Param("id"))
	if product.ID == 0 {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	h := DefaultH(c)
	h["Title"] = "Редактирование карточки изделия"
	h["Product"] = product
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "products/form", h)
}

//ProductUpdate handles POST /admin/products/:id/edit route
func ProductUpdate(c *gin.Context) {
	product := models.Product{}
	db := models.GetDB()
	if err := c.ShouldBind(&product); err != nil {
		session := sessions.Default(c)
		session.AddFlash(err.Error())
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/products")
		return
	}
	db.Where("id in (?)", product.ImageIds).Find(&product.Images)
	if err := db.Save(&product).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/products")
}

//ProductDelete handles POST /admin/products/:id/delete route
func ProductDelete(c *gin.Context) {
	product := models.Product{}
	db := models.GetDB()
	db.First(&product, c.Param("id"))
	if product.ID == 0 {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	if err := db.Delete(&product).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/products")
}
