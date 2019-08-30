package controllers

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/denisbakhtin/panteon/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

//AdvantageIndex handles GET /admin/advantages route
func AdvantageIndex(c *gin.Context) {
	db := models.GetDB()
	var advantages []models.Advantage
	db.Order("id").Find(&advantages)
	h := DefaultH(c)
	h["Title"] = "Наши преимущества"
	h["Advantages"] = advantages
	c.HTML(http.StatusOK, "advantages/index", h)
}

//AdvantageNew handles GET /admin/new_advantage route
func AdvantageNew(c *gin.Context) {
	h := DefaultH(c)
	h["Title"] = "Преимущество"
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	h["Advantage"] = models.Advantage{}
	session.Save()

	c.HTML(http.StatusOK, "advantages/form", h)
}

//AdvantageCreate handles POST /admin/new_advantage route
func AdvantageCreate(c *gin.Context) {
	db := models.GetDB()
	advantage := models.Advantage{}
	if err := c.ShouldBind(&advantage); err != nil {
		session := sessions.Default(c)
		session.AddFlash(err.Error())
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/new_advantage")
		return
	}

	if err := db.Create(&advantage).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/advantages")
}

//AdvantageEdit handles GET /admin/advantages/:id/edit route
func AdvantageEdit(c *gin.Context) {
	db := models.GetDB()
	advantage := models.Advantage{}
	db.First(&advantage, c.Param("id"))
	if advantage.ID == 0 {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	h := DefaultH(c)
	h["Title"] = "Редактирование преимущества"
	h["Advantage"] = advantage
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "advantages/form", h)
}

//AdvantageUpdate handles POST /admin/advantages/:id/edit route
func AdvantageUpdate(c *gin.Context) {
	advantage := models.Advantage{}
	db := models.GetDB()
	if err := c.ShouldBind(&advantage); err != nil {
		session := sessions.Default(c)
		session.AddFlash(err.Error())
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/advantages")
		return
	}
	if err := db.Save(&advantage).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/advantages")
}

//AdvantageDelete handles POST /admin/advantages/:id/delete route
func AdvantageDelete(c *gin.Context) {
	advantage := models.Advantage{}
	db := models.GetDB()
	db.First(&advantage, c.Param("id"))
	if advantage.ID == 0 {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	if err := db.Delete(&advantage).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/advantages")
}
