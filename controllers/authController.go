package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/sanjiv-madhavan/go-jwt-auth/cache"
	"github.com/sanjiv-madhavan/go-jwt-auth/constants"
	"github.com/sanjiv-madhavan/go-jwt-auth/database"
	"github.com/sanjiv-madhavan/go-jwt-auth/models"
	"github.com/sanjiv-madhavan/go-jwt-auth/utils"
)

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	var user models.User
	var foundUser models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		c.logger.Error("Invalid request body")
		c.middleware.SendJSONResponse(w, http.StatusBadRequest, "User data invalid")
		return
	}
	defer cancel()
	userCollection := database.OpenCollection(*database.NewMongoClient(r.Context(), c.logger), "user")
	err := userCollection.FindOne(r.Context(), bson.M{"email": user.Email}).Decode(&foundUser)
	if err != nil {
		c.logger.Error("User mail or password incorrect")
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "User mail or password incorrect")
		return
	}
	passwordValid, err := c.VerifyPasswords(user.Password, foundUser.Password)
	if err != nil {
		c.logger.Error("Password validation failed")
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "Password validation failed")
		return
	}
	if !passwordValid {
		c.logger.Error("Password invalid")
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "Password invalid")
		return
	}

	token, refreshToken, err := utils.GenerateAllTokens(foundUser.Email, foundUser.FirstName, foundUser.LastName, foundUser.UserType, foundUser.UID)
	if err != nil {
		c.logger.Error("Failed to generate token", slog.Any("Error: ", err))
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	if err = utils.UpdateAllTokens(r.Context(), c.logger, token, refreshToken, foundUser.UID); err != nil {
		c.logger.Error("Failed to update token", slog.Any("Error: ", err))
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "Failed to update token")
		return
	}

	if err = userCollection.FindOne(r.Context(), bson.M{"user_id": foundUser.UID}).Decode(&foundUser); err != nil {
		c.logger.Error("Unable to find user", slog.Any("Error: ", err))
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "Unable to find user")
		return
	}

	c.logger.Info(fmt.Sprintf("User %s logged in", foundUser.UID))
	c.middleware.SendJSONResponse(w, http.StatusOK, foundUser)
}

func (c *Controller) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.logger.Error(http.StatusText(http.StatusInternalServerError))
		return "", err
	}
	return string(bytes), nil
}

func (c *Controller) VerifyPasswords(givenPassword string, userPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(givenPassword))

	if err != nil {
		c.logger.Error("mail or password incorrect", slog.Any("Error: ", err))
		return false, err
	}

	return true, nil
}

func (c *Controller) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var passwordUpdateRequest models.PasswordUpdateRequest
	var user models.User
	userID := r.Context().Value(constants.UID).(string)
	userMail := r.Context().Value(constants.Email).(string)
	if err := utils.MatchUsertoID(w, r, userID); err != nil {
		c.logger.Error("User unauthorized")
		c.middleware.SendJSONResponse(w, http.StatusUnauthorized, "User unauthorized")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&passwordUpdateRequest); err != nil {
		c.logger.Error("Invalid request", slog.Any("Error: ", err))
		c.middleware.SendJSONResponse(w, http.StatusBadRequest, "invalid request")
		return
	}

	userCollection := database.OpenCollection(*database.NewMongoClient(ctx, c.logger), "user")
	if err := userCollection.FindOne(ctx, bson.M{"email": userMail}).Decode(&user); err != nil {
		c.logger.Error("User not found", slog.Any("Error:", err))
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "User not found")
		return
	}

	if ok, err := c.VerifyPasswords(passwordUpdateRequest.OldPassword, user.Password); !ok {
		c.logger.Error("Old password incorrect", slog.Any("Error:", err))
		c.middleware.SendJSONResponse(w, http.StatusBadRequest, "Old password incorrect")
		return
	}

	newPasswordHash, err := c.HashPassword(passwordUpdateRequest.NewPassword)
	if err != nil {
		c.logger.Error("Unable to store password")
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "Unable to store password")
	}
	updateObject := bson.M{"$set": bson.M{"password": newPasswordHash}}
	if _, err := userCollection.UpdateOne(ctx, bson.M{"email": user.Email}, updateObject); err != nil {
		c.logger.Error("Failed to update password", slog.Any("error", err))
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	// To Delete - Creation of redis client - make it singleton
	redisClient := cache.NewRedisClient(c.logger)
	ttl := r.Context().Value(constants.ExpiresAt).(int64)
	cache.SetUserSpecificInvalidation(r.Context(), redisClient, userID, time.Now().Unix(), ttl)

	c.logger.Info("Password updated successfully")
	c.middleware.SendJSONResponse(w, http.StatusOK, "Password updated successfully")
}

func (c *Controller) Signup(w http.ResponseWriter, r *http.Request) {
	// To Delete - too many timeouts configured at a granular level
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var user models.User
	r = r.WithContext(ctx)
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		c.logger.Error("Invalid Request Body", slog.Any("Error: ", err))
		c.middleware.SendJSONResponse(w, http.StatusBadRequest, "Invalid User data")
		return
	}
	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		c.logger.Error("Invalid Request Body", slog.Any("Error: ", err))
		c.middleware.SendJSONResponse(w, http.StatusBadRequest, "Invalid User data")
		return
	}

	// To Delete - too many instances of userCollection
	userCollection := database.OpenCollection(*database.NewMongoClient(r.Context(), c.logger), "user")
	count, err := userCollection.CountDocuments(r.Context(), bson.M{"email": user.Email})
	if err != nil {
		c.logger.Error(http.StatusText(http.StatusInternalServerError))
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	if count > 0 {
		c.logger.Error(fmt.Sprintf("user with the mail ID %s already exists", user.Email))
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("user with the mail ID %s already exists", user.Email))
		return
	}

	user.Created_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	if err != nil {
		c.logger.Error(http.StatusText(http.StatusInternalServerError))
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	user.Updated_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	if err != nil {
		c.logger.Error(http.StatusText(http.StatusInternalServerError))
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	user.ID = primitive.NewObjectID()
	user.UID = user.ID.Hex()

	password, err := c.HashPassword(user.Password)
	if err != nil {
		c.logger.Error("Unable to store password")
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "Unable to store password")
	}
	user.Password = password

	token, refreshToken, err := utils.GenerateAllTokens(user.Email, user.FirstName, user.LastName, user.UserType, user.UID)
	if err != nil {
		c.logger.Error("Failed to generate token", slog.Any("Error: ", err))
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}
	user.Token = token
	user.RefreshToken = refreshToken

	resultInsertionNumber, err := userCollection.InsertOne(r.Context(), user)
	if err != nil {
		c.logger.Error("User insertion failed", slog.Any("Error: ", err))
		// To Delete - give detailed error above and summarize below. create separate file for below ones
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "User insertion failed")
		return
	}

	c.logger.Info("User signed up")
	c.middleware.SendJSONResponse(w, http.StatusOK, resultInsertionNumber)
}
