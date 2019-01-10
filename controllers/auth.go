package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/denisbakhtin/panteon/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

//SignInGet handles GET /signin route
func SignInGet(c *gin.Context) {
	h := DefaultH(c)
	h["Title"] = "Вход в систему"
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "auth/signin", h)
}

//SignInPost handles POST /signin route, authenticates user
func SignInPost(c *gin.Context) {
	session := sessions.Default(c)
	login := models.Login{}
	db := models.GetDB()
	returnURL := c.DefaultQuery("return", "/admin")
	if err := c.ShouldBind(&login); err != nil {
		session.AddFlash("Пожалуйста, укажите правильные данные.")
		session.Save()
		c.Redirect(http.StatusFound, "/signin")
		return
	}

	user := models.User{}
	db.Where("email = lower(?)", login.Email).First(&user)

	if user.ID == 0 || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)) != nil {
		logrus.Errorf("Login error, IP: %s, Email: %s", c.ClientIP(), login.Email)
		session.AddFlash("Электронная почта или пароль указаны неверно")
		session.Save()
		c.Redirect(http.StatusFound, fmt.Sprintf("/signin?return=%s", url.QueryEscape(returnURL)))
		return
	}

	session.Set(userIDKey, user.ID)
	session.Save()
	c.Redirect(http.StatusFound, returnURL)
}

//SignUpGet handles GET /signup route
func SignUpGet(c *gin.Context) {
	h := DefaultH(c)
	h["Title"] = "Регистрация в системе"
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "auth/signup", h)
}

//SignUpPost handles POST /signup route, creates new user
func SignUpPost(c *gin.Context) {
	session := sessions.Default(c)
	register := models.Register{}
	db := models.GetDB()
	if err := c.ShouldBind(&register); err != nil {
		session.AddFlash("Пожалуйста, заполните все обязательные поля.")
		session.Save()
		c.Redirect(http.StatusFound, "/signup")
		return
	}
	if register.Password != register.PasswordConfirm {
		session.AddFlash("Пароль и подтверждение пароля не совпадают.")
		session.Save()
		c.Redirect(http.StatusFound, "/signup")
		return
	}
	register.Email = strings.ToLower(register.Email)
	user := models.User{}
	db.Where("email = ?", register.Email).First(&user)
	if user.ID != 0 {
		session.AddFlash("Пользователь с такой электронной почтой уже существует")
		session.Save()
		c.Redirect(http.StatusFound, "/signup")
		return
	}
	//create user
	user.Email = register.Email
	user.Password = register.Password
	user.FirstName = register.FirstName
	user.MiddleName = register.MiddleName
	user.LastName = register.LastName
	user.Role = models.MEMBER
	if err := db.Create(&user).Error; err != nil {
		session.AddFlash("Ошибка регистрации нового пользователя.")
		session.Save()
		logrus.Errorf("Error whilst registering user: %v", err)
		c.Redirect(http.StatusFound, "/signup")
		return
	}

	session.Set(userIDKey, user.ID)
	session.Save()
	c.Redirect(http.StatusFound, "/")
	return
}

//SignoutGet handles GET /signout route
func SignoutGet(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(userIDKey)
	session.Save()
	c.Redirect(http.StatusSeeOther, "/")
}
