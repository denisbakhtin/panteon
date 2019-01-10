package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/denisbakhtin/panteon/config"
	"github.com/denisbakhtin/panteon/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	gomail "gopkg.in/gomail.v2"
)

//OrderGet handles GET /admin/orders/:id route
func OrderGet(c *gin.Context) {
	db := models.GetDB()
	order := models.Order{}
	db.Preload("Products").First(&order, c.Param("id"))
	if order.ID == 0 {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	h := DefaultH(c)
	h["Title"] = fmt.Sprintf("Заказ №%d", order.ID)
	h["Order"] = order
	c.HTML(http.StatusOK, "orders/show", h)
}

//OrderIndex handles GET /admin/orders route
func OrderIndex(c *gin.Context) {
	db := models.GetDB()
	var orders []models.Order
	db.Order("id").Find(&orders)
	h := DefaultH(c)
	h["Title"] = "Список заказов"
	h["Orders"] = orders
	c.HTML(http.StatusOK, "orders/index", h)
}

//OrderNew handles GET /new_order route
func OrderNew(c *gin.Context) {
	h := DefaultH(c)
	h["Title"] = "Оформление заказа"
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	h["Order"] = models.Order{}
	session.Save()

	c.HTML(http.StatusOK, "orders/form", h)
}

//OrderCreate handles POST /new_order route
func OrderCreate(c *gin.Context) {
	db := models.GetDB()
	order := models.Order{}
	session := sessions.Default(c)
	if err := c.ShouldBind(&order); err != nil {
		session.AddFlash(err.Error())
		session.Save()
		c.Redirect(http.StatusSeeOther, "/new_order")
		return
	}

	cart := getCart(c)
	if len(cart) == 0 {
		session.AddFlash("Оформление заказа невозможно, т.к. ваша корзина пуста.")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/new_order")
		return
	}
	db.Where("id in(?)", getCartProductIDs(cart)).Find(&order.Products)

	if err := db.Create(&order).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	session.Delete("cart")
	session.Save()

	//send email notifications
	notifyAdminOfOrder(c, &order)
	notifyClientOfOrder(c, &order)

	c.Redirect(http.StatusFound, fmt.Sprintf("/confirm_order/%d", order.ID))
}

//OrderConfirm handles GET /confirm_order/:id route
func OrderConfirm(c *gin.Context) {
	db := models.GetDB()
	order := models.Order{}
	db.First(&order, c.Param("id"))
	h := DefaultH(c)
	h["Title"] = "Ваш заказ принят"
	h["Order"] = order
	c.HTML(http.StatusOK, "orders/confirm", h)
}

//OrderDelete handles POST /admin/orders/:id/delete route
func OrderDelete(c *gin.Context) {
	order := models.Order{}
	db := models.GetDB()
	db.First(&order, c.Param("id"))
	if order.ID == 0 {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	if err := db.Delete(&order).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/orders")
}

func notifyAdminOfOrder(c *gin.Context, order *models.Order) {
	//closure is needed here, as r may be released by the time func finishes
	go func() {
		var b bytes.Buffer

		tmpl := template.New("").Funcs(getFuncMap())
		workingdir, _ := os.Getwd()
		tmpl, _ = tmpl.ParseFiles(path.Join(workingdir, "views", "emails", "admin_order.gohtml"))
		if err := tmpl.Lookup("emails/admin_order").Execute(&b, gin.H{"Order": order}); err != nil {
			logrus.Error(err)
			return
		}

		smtp := config.GetConfig().SMTP
		msg := gomail.NewMessage()
		msg.SetHeader("From", smtp.From)
		msg.SetHeader("To", getSetting("order_email"))
		msg.SetHeader("Subject", "Заказ на сайте www.panteon-vlz.ru")
		msg.SetBody(
			"text/html",
			b.String(),
		)

		port, _ := strconv.Atoi(smtp.Port)
		dialer := gomail.NewPlainDialer(smtp.SMTP, port, smtp.User, smtp.Password)
		sender, err := dialer.Dial()
		if err != nil {
			logrus.Error(err)
			return
		}
		if err := gomail.Send(sender, msg); err != nil {
			logrus.Error(err)
			return
		}
	}()
}

func notifyClientOfOrder(c *gin.Context, order *models.Order) {
	//closure is needed here, as r may be released by the time func finishes
	go func() {
		var b bytes.Buffer

		tmpl := template.New("").Funcs(getFuncMap())
		workingdir, _ := os.Getwd()
		tmpl, _ = tmpl.ParseFiles(path.Join(workingdir, "views", "emails", "order.gohtml"))
		if err := tmpl.Lookup("emails/order").Execute(&b, gin.H{"Order": order}); err != nil {
			logrus.Error(err)
			return
		}

		smtp := config.GetConfig().SMTP
		msg := gomail.NewMessage()
		msg.SetHeader("From", smtp.From)
		msg.SetHeader("To", order.Email)
		msg.SetHeader("Subject", "Заказ на сайте www.panteon-vlz.ru")
		msg.SetBody(
			"text/html",
			b.String(),
		)

		port, _ := strconv.Atoi(smtp.Port)
		dialer := gomail.NewPlainDialer(smtp.SMTP, port, smtp.User, smtp.Password)
		sender, err := dialer.Dial()
		if err != nil {
			logrus.Error(err)
			return
		}
		if err := gomail.Send(sender, msg); err != nil {
			logrus.Error(err)
			return
		}
	}()
}
