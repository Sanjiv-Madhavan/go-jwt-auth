package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id"`
	FirstName    string             `json:"first_name" validate:"required,min=2,max=30"`
	LastName     string             `json:"last_name" validate:"required,min=2,max=30"`
	Email        string             `json:"email" validate:"email,required"`
	Password     string             `json:"password" validate:"required"`
	UserType     string             `json:"user_type" validate:"required,eq=ADMIN|eq=USER"`
	Token        string             `json:"token"`
	RefreshToken string             `json:"refresh_token"`
	Created_at   time.Time          `json:"created_at"`
	Updated_at   time.Time          `json:"updated_at"`
	UID          string             `json:"uid"`
}

type PasswordUpdateRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}
