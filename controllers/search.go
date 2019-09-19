package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/denisbakhtin/panteon/models"
	"github.com/gin-gonic/gin"
)

//SearchGet handles GET /search?search= route
func SearchGet(c *gin.Context) {
	limit := 48
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	db := models.GetDB()
	var products []models.Product

	dbq := db.Model(&models.Product{}).Where("published = true AND lower(title) LIKE lower(?)", fmt.Sprintf("%%%s%%", strings.TrimSpace(c.Query("search"))))
	paginator := buildPaginator(dbq, c.Request.URL.Path, c.Request.URL.RawQuery, limit, page)
	dbq.Offset((page - 1) * limit).Limit(limit).
		Order("recommended desc, id asc").Preload("Images").Find(&products)

	h := DefaultH(c)
	h["Title"] = fmt.Sprintf("Результаты поиска по запросу: %s", c.Query("search"))
	h["Products"] = products
	h["Paginator"] = paginator
	c.HTML(http.StatusOK, "search/show", h)
}
