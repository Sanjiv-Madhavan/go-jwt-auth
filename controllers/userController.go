package controllers

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sanjiv-madhavan/go-jwt-auth/constants"
	"github.com/sanjiv-madhavan/go-jwt-auth/database"
	"github.com/sanjiv-madhavan/go-jwt-auth/models"
	"github.com/sanjiv-madhavan/go-jwt-auth/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (c *Controller) ListUsers(w http.ResponseWriter, r *http.Request) {
	if err := utils.CheckUserType(w, r, "ADMIN"); err != nil {
		c.logger.Error("User unauthorized", slog.Any("err: ", err))
		c.middleware.SendJSONResponse(w, http.StatusBadRequest, "User unauthorized")
		return
	}
	var ctx, cancel = context.WithTimeout(r.Context(), 100*time.Second)
	defer cancel()

	userCollection := database.OpenCollection(*database.NewMongoClient(r.Context(), c.logger), "user")
	recordPerPage := 10
	page := 1
	startIndex := 0

	queryParams := r.URL.Query()
	if val, ok := queryParams["recordPerPage"]; ok {
		if recordPerPageValue, err := strconv.Atoi(val[0]); err == nil && recordPerPageValue > 0 {
			recordPerPage = recordPerPageValue
		}
	}
	if val, ok := queryParams["page"]; ok {
		if pg, err := strconv.Atoi(val[0]); err == nil && pg > 0 {
			page = pg
		}
	}

	startIndex = (page - 1) * recordPerPage
	if val, ok := queryParams["startIndex"]; ok {
		if si, err := strconv.Atoi(val[0]); err == nil {
			startIndex = si
		}
	}

	matchStage := bson.D{{"$match", bson.D{{}}}}
	groupStage := bson.D{{"$group", bson.D{
		{"_id", bson.D{{"_id", "null"}}},
		{"total_count", bson.D{{"$sum", 1}}},
		{"data", bson.D{{"$push", "$$ROOT"}}}}}}
	projectStage := bson.D{
		{"$project", bson.D{
			{"_id", 0},
			{"total_count", 1},
			{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}}}}}
	result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage, groupStage, projectStage})
	if err != nil {
		c.logger.Error("error occured while listing user items", slog.Any("err: ", err))
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "error occured while listing user items")
	}
	var allusers []bson.M
	if err = result.All(ctx, &allusers); err != nil {
		log.Fatal(err)
	}
	c.middleware.SendJSONResponse(w, http.StatusOK, allusers[0])
	c.logger.Info("Users found")
}

func (c *Controller) ListUserById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params[constants.ParamUserID]
	// params userID must match with the auth claims uid, the user_type must be "USER" or "ADMIN"
	if err := utils.MatchUsertoID(w, r, userID); err != nil {
		c.logger.Error("Unable to match the user to the database", slog.Any("Error: ", err))
		c.middleware.SendJSONResponse(w, http.StatusBadRequest, "User given does not map to users in the database")
		return
	}
	var user models.User
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	r = r.WithContext(ctx)
	userCollection := database.OpenCollection(*database.NewMongoClient(r.Context(), c.logger), "user")
	err := userCollection.FindOne(r.Context(), bson.M{"uid": userID}).Decode(&user)
	if err != nil {
		c.logger.Error(fmt.Sprintf("Unable to parse user %s", userID))
		c.middleware.SendJSONResponse(w, http.StatusInternalServerError, "Check the user data")
		return
	}
	c.middleware.SendJSONResponse(w, http.StatusOK, user)
}
