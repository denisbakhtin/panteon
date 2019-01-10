package models

import (
	"golang.org/x/crypto/bcrypt"
)

//MEMBER is a member role name
const MEMBER = "member"

//ADMIN is an admin role name
const ADMIN = "admin"

//ANONYMOUS is an anonymous role name
const ANONYMOUS = "anonymous"

//Login view model
type Login struct {
	Email    string `form:"email" binding:"required"`
	Password string `form:"password" binding:"required"`
}

//Register view model
type Register struct {
	FirstName       string `form:"first_name" binding:"required"`
	MiddleName      string `form:"middle_name" binding:"required"`
	LastName        string `form:"last_name" binding:"required"`
	Email           string `form:"email" binding:"required"`
	Password        string `form:"password" binding:"required"`
	PasswordConfirm string `form:"password_confirm" binding:"required"`
}

//Manage user view model
type Manage struct {
	Name            string `form:"name" binding:"required"`
	Email           string `form:"email" binding:"required"`
	Password        string `form:"password" binding:"required"`
	PasswordConfirm string `form:"password_confirm" binding:"required"`
	NewPassword     string `form:"new_password" binding:"required"`
}

//User type contains user info
type User struct {
	Model

	Email      string `form:"email" binding:"required"`
	FirstName  string `form:"first_name"`
	MiddleName string `form:"middle_name"`
	LastName   string `form:"last_name"`
	Password   string `form:"password" binding:"required"`
	Role       string `form:"role" binding:"required"`
}

//BeforeSave gorm hook
func (u *User) BeforeSave() (err error) {
	var hash []byte
	hash, err = bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	u.Password = string(hash)
	return
}

//IsAdmin checks if user is admin
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

//IsMember checks if user is a member
func (u *User) IsMember() bool {
	return u.Role == "member"
}
