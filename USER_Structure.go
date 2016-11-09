// Contains the object structure and methods relating to the USER.
package main

import (
	"encoding/json"

	"github.com/Esseh/retrievable"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

var (
	UsersTable          = "Users"
	RecentlyViewedTable = "RecentlyViewedCourses"
)

type (
	User struct {
		First, Last       string
		Email             string
		Avatar            bool `datastore:",noindex"`
		Bio               string
		retrievable.IntID `datastore:"-" json:"-"`
	}
	// RecentlyViewed struct{ CID []int64 } // Holds recently viewed course information.
	EncryptedUser struct {
		First, Last string
		Email       string
		Avatar      bool `datastore:",noindex"`
		Bio         string
	}
)

// int64 Keys
func (u *User) Key(ctx context.Context, key interface{}) *datastore.Key {
	if v, ok := key.(retrievable.IntID); ok {
		return datastore.NewKey(ctx, UsersTable, "", int64(v), nil)
	}
	return datastore.NewKey(ctx, UsersTable, "", key.(int64), nil)
}

// func (r *RecentlyViewed) Key(ctx context.Context, key interface{}) *datastore.Key {
// 	return datastore.NewKey(ctx, RecentlyViewedTable, "", key.(int64), nil)
// }

func (u *User) toEncrypt() (*EncryptedUser, error) {
	e := &EncryptedUser{
		First:     u.First,
		Last:      u.Last,
		Avatar:    u.Avatar,
		Bio:       u.Bio,
	}
	email, err := Encrypt([]byte(u.Email), encryptKey)
	if err != nil {
		return nil, err
	}
	e.Email = email
	return e, nil
}

func (u *User) fromEncrypt(e *EncryptedUser) error {
	email, err := Decrypt(e.Email, encryptKey)
	if err != nil {
		return err
	}
	u.First = e.First
	u.Last = e.Last
	u.Email = string(email)
	u.Avatar = e.Avatar
	u.Bio = e.Bio
	return nil
}

func (u *User) Serialize() []byte {
	data, _ := u.toEncrypt()
	ret, _ := json.Marshal(&data)
	return ret
}

func (u *User) Unserialize(data []byte) error {
	e := &EncryptedUser{}
	err := json.Unmarshal(data, e)
	if err != nil {
		return err
	}
	return u.fromEncrypt(e)
}

func (u *User) Save() ([]datastore.Property, error) {
	e, err := u.toEncrypt()
	if err != nil {
		return nil, err
	}
	return datastore.SaveStruct(e)
}

func (u *User) Load(ps []datastore.Property) error {
	e := &EncryptedUser{}
	err := datastore.LoadStruct(e, ps)
	if err != nil {
		return err
	}
	return u.fromEncrypt(e)
}
