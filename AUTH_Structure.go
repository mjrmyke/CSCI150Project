/// Contains the structure for our logins/user structure.
package main

import (
	"github.com/Esseh/retrievable"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"time"
)

const (
	LoginTable         = "Login"
	SessionTable       = "Session"
	PasswordResetTable = "PasswordReset"
)

type (
	// Login, leads to a temporary session
	LoginLocalAccount struct {
		UserID   int64
		Password []byte
	}
	// A temporary session, can retrieve user information.
	Session struct {
		UserID       int64
		ID           int64  `datastore:"-"`
		IP           string `datastore:",noindex"`
		BrowserUsed  string `datastore:",noindex"`
		LocationUsed string `datastore:",noindex"`
		LastUsed     time.Time
	}
	// A one time key that can be used to reset the password of a specific user id.
	PasswordReset struct {
		UID      int64
		Creation time.Time // If it's too old it can be deleted
	}
)

// These two functions are used in place of retrievable, if LoginLocal is not appropriate it will bypass and use OAuth instead
func (l *LoginLocalAccount) Get(ctx context.Context, key interface{}) error {
	getErr := retrievable.GetEntity(ctx, key, l) // LoginLocal Case
	if getErr != nil {                           // OAuth Case
		oauth := LoginOauthAccount{}
		ogetErr := retrievable.GetEntity(ctx, key, &oauth)
		if ogetErr != nil {
			return ogetErr
		}
		l.UserID = oauth.UserID
	}
	return nil
}
func (l *LoginLocalAccount) Place(ctx context.Context, key interface{}) (*datastore.Key, error) {
	if string(l.Password) == "" { // OAuth Case
		oauth := LoginOauthAccount{}
		oauth.UserID = oauth.UserID
		return retrievable.PlaceEntity(ctx, key, &oauth)
	} else { // LoginLocal Case
		return retrievable.PlaceEntity(ctx, key, l)
	}
}

// String Keys
func (l *LoginLocalAccount) Key(ctx context.Context, key interface{}) *datastore.Key {
	e, _ := Encrypt([]byte(key.(string)), encryptKey)
	return datastore.NewKey(ctx, LoginTable, e, 0, nil)
}

// int64 Keys
func (p PasswordReset) Key(ctx context.Context, key interface{}) *datastore.Key {
	return datastore.NewKey(ctx, PasswordResetTable, "", key.(int64), nil)
}
func (s *Session) Key(ctx context.Context, key interface{}) *datastore.Key {
	return datastore.NewKey(ctx, SessionTable, "", key.(int64), nil)
}

// StoreKey
func (s *Session) StoreKey(key *datastore.Key) { s.ID = key.IntID() }
