// Contains the object structure and methods relating to the USER.
package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Esseh/retrievable"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
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
		DOB               time.Time
		Customer          string `datastore:",noindex"`
		Notified          int64
		Questions         int64
		retrievable.IntID `datastore:"-" json:"-"`
	}
	RecentlyViewed struct{ CID []int64 } // Holds recently viewed course information.
	EncryptedUser  struct {
		First, Last string
		Email       string
		Avatar      bool `datastore:",noindex"`
		Bio         string
		DOB         time.Time
		Customer    string `datastore:",noindex"`
		Notified    int64
		Questions   int64
	}
)

// int64 Keys
func (u *User) Key(ctx context.Context, key interface{}) *datastore.Key {
	return datastore.NewKey(ctx, UsersTable, "", int64(key.(retrievable.IntID)), nil)
}
func (r *RecentlyViewed) Key(ctx context.Context, key interface{}) *datastore.Key {
	return datastore.NewKey(ctx, RecentlyViewedTable, "", key.(int64), nil)
}

func (u *User) toEncrypt() (*EncryptedUser, error) {
	e := &EncryptedUser{
		First:     u.First,
		Last:      u.Last,
		Avatar:    u.Avatar,
		Bio:       u.Bio,
		DOB:       u.DOB,
		Customer:  u.Customer,
		Notified:  u.Notified,
		Questions: u.Questions,
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
	u.DOB = e.DOB
	u.Customer = e.Customer
	u.Notified = e.Notified
	u.Questions = e.Questions
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

// Gets the user's most recently viewed courses.
func (r *RecentlyViewed) Get(req *http.Request) (context.Context, int64, error) {
	u, _ := GetUserFromSession(req)
	ctx := appengine.NewContext(req)
	getErr := retrievable.GetEntity(ctx, u.IntID, r)
	if getErr != nil {
		_, placeErr := retrievable.PlaceEntity(ctx, u.IntID, r)
		if placeErr != nil {
			return ctx, int64(u.IntID), placeErr
		}
	}
	return ctx, int64(u.IntID), nil
}

// Updates the user's most recently viewed courses.
func (r *RecentlyViewed) Update(req *http.Request, p httprouter.Params) error {
	ctx, k, err := r.Get(req)
	if err != nil {
		return err
	}

	v, valErr := strconv.ParseInt(p.ByName("ID"), 10, 64)
	if valErr != nil {
		return valErr
	}
	if len(r.CID) > 7 {
		r.CID = append([]int64{v}, r.CID[:len(r.CID)-2]...)
	} else {
		r.CID = append([]int64{v}, r.CID...)
	}

	_, placeErr := retrievable.PlaceEntity(ctx, k, r)
	if placeErr != nil {
		return placeErr
	}
	return nil
}
