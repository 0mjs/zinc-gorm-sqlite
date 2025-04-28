package main

import (
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type CreateUserRequest struct {
	Name string `json:"name"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type DeleteUserRequest struct {
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}
