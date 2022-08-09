package main

import (
	"github.com/Shitomo/play-kafka-chat-core/model"
	"github.com/gorilla/websocket"
)

type websocketConn struct {
	*websocket.Conn
}

type UserID int64

type User struct {
	id   UserID
	conn websocketConn
}

func NewUser(id UserID, conn *websocket.Conn) User {
	return User{
		id:   id,
		conn: websocketConn{conn},
	}
}

type Users struct {
	users map[UserID]*User
}

func NewUsers() Users {
	return Users{
		map[UserID]*User{},
	}
}

func (u Users) Contains(user User) bool {
	val := u.users[user.id]
	return val != nil
}

func (u *Users) Append(user User) {
	if u.Contains(user) {
		return
	}
	u.users[user.id] = &user
}

func (u Users) Send(id UserID, message model.Message) error {
	target := u.users[id]
	if target == nil {
		return UserNotFoundError{UserID: id}
	}

	return target.conn.WriteJSON(message)
}
