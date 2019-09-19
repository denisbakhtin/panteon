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

//OrderPost handles POST /order route
func OrderPost(c *gin.Context) {
	order := models.Order{}
	session := sessions.Default(c)
	if err := c.ShouldBind(&order); err != nil {
		session.AddFlash(err.Error())
		session.Save()
		c.Redirect(http.StatusSeeOther, order.BackURL)
		return
	}

	models.GetDB().First(&order.Product, order.ProductID)
	//send email notifications
	notifyAdminOfOrder(c, &order)
	notifyClientOfOrder(c, &order)

	session.AddFlash("Ваш запрос принят.")
	session.Save()
	c.Redirect(http.StatusFound, order.BackURL)
}

func notifyAdminOfOrder(c *gin.Context, order *models.Order) {
	//closure is needed here, as r may be released by the time func finishes
	go func() {
		var b bytes.Buffer

		domain := config.GetConfig().Domain
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
		msg.SetHeader("Subject", fmt.Sprintf("Запрос на сайте %s", domain))
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

		domain := config.GetConfig().Domain
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
		msg.SetHeader("Subject", fmt.Sprintf("Запрос на сайте %s", domain))
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
