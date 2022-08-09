package main

import "fmt"

type UserNotFoundError struct {
	UserID UserID
}

func (u UserNotFoundError) Error() string {
	return fmt.Sprintf("user id %d is not found", u.UserID)
}

func NewUserNotFoundError(userID UserID) UserNotFoundError {
	return UserNotFoundError{
		UserID: userID,
	}
}
