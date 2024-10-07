package utils

import (
	"errors"
	"net/http"

	"github.com/sanjiv-madhavan/go-jwt-auth/constants"
)

func CheckUserType(w http.ResponseWriter, r *http.Request, role string) (err error) {
	userType := r.Context().Value(constants.UserType)
	err = nil

	if userType != role {
		err = errors.New("unauthorized to access the resource")
		return err
	}

	return err
}

func MatchUsertoID(w http.ResponseWriter, r *http.Request, userID string) (err error) {
	userType := r.Context().Value(constants.UserType).(string)
	uid := r.Context().Value(constants.UID).(string)
	err = nil

	if userType == "USER" && uid != userID {
		err = errors.New("unauthorized to access the resource")
		return err
	}

	// Handle admin and other users
	err = CheckUserType(w, r, userType)
	return err
}
