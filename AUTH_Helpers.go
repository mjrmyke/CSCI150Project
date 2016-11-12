// Contains helper functions for dealing with authentication
package main

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Esseh/retrievable"
	"github.com/mssola/user_agent"
	"github.com/pariz/gountries"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

const sessionTime = 7 * 24 * 60 * 60

//=============================================================================================
func GetUserIDFromLogin(ctx context.Context, email, password string) (int64, error) {
	urID := LoginLocalAccount{}
	if getErr := retrievable.GetEntity(ctx, email, &urID); getErr != nil {
		return -1, getErr
	}

	if compareErr := bcrypt.CompareHashAndPassword(urID.Password, []byte(password)); compareErr != nil {
		return -1, compareErr
	}

	return urID.UserID, nil
}

//=============================================================================================
func CreateUserFromLogin(ctx context.Context, email, password string, u *User) (*User, error) {
	checkLogin := LoginLocalAccount{}

	// Check that user does not exist
	if checkErr := retrievable.GetEntity(ctx, email, &checkLogin); checkErr == nil {
		return u, ErrUsernameExists
	} else if checkErr != datastore.ErrNoSuchEntity && checkErr != nil {
		return u, checkErr
	}

	ukey, putUserErr := retrievable.PlaceEntity(ctx, retrievable.IntID(0), u)
	if putUserErr != nil {
		return u, putUserErr
	}

	if u.IntID == 0 {
		return u, errors.New("HEY, DATASTORE IS STUPID")
	}

	// Set Password
	cryptPass, cryptErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if cryptErr != nil {
		return u, cryptErr
	}

	uLogin := LoginLocalAccount{
		Password: cryptPass,
		UserID:   ukey.IntID(),
	}
	_, putErr := retrievable.PlaceEntity(ctx, email, &uLogin)
	return u, putErr
}

//=============================================================================================
func DeleteUserIDAndLoginKey(ctx context.Context, username, password string) error {
	usrID, getErr := GetUserIDFromLogin(ctx, username, password)
	if getErr != nil {
		return getErr
	}

	if usrDeleteErr := retrievable.DeleteEntity(ctx, (&User{}).Key(ctx, usrID)); usrDeleteErr != nil {
		return usrDeleteErr
	}
	return retrievable.DeleteEntity(ctx, (&LoginLocalAccount{}).Key(ctx, username))
}

//=============================================================================================
func ChangePassword(ctx context.Context, email, oldPassword, newPassword string) error {
	if isValid, validErr := ValidatePassword(newPassword); !isValid {
		return validErr
	}

	loginAccount := LoginLocalAccount{}
	if getErr := retrievable.GetEntity(ctx, email, &loginAccount); getErr != nil {
		return getErr
	}

	if compErr := bcrypt.CompareHashAndPassword(loginAccount.Password, []byte(oldPassword)); compErr != nil {
		return compErr
	}

	cryptPass, cryptErr := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if cryptErr != nil {
		return cryptErr
	}

	loginAccount.Password = cryptPass
	if _, placeErr := retrievable.PlaceEntity(ctx, email, &loginAccount); placeErr != nil {
		return placeErr
	}

	return nil
}

//=============================================================================================
func ChangeEmail(ctx context.Context, oldEmail, newEmail string) error {
	loginAccount := LoginLocalAccount{}
	if getErr := retrievable.GetEntity(ctx, oldEmail, &loginAccount); getErr != nil {
		return getErr
	}

	oldKey := loginAccount.Key(ctx, oldEmail)
	if deleteErr := retrievable.DeleteEntity(ctx, oldKey); deleteErr != nil {
		return deleteErr
	}

	if _, placeErr := retrievable.PlaceEntity(ctx, newEmail, &loginAccount); placeErr != nil {
		return placeErr
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////////
// will check if a password follows the required protocols for validation.
// TODO: Implement, Low Priority
///////////////////////////////////////////////////////////////////////////////////
func ValidatePassword(password string) (bool, error) {
	return true, ErrNotImplemented
}

//=============================================================================================
// Initializes a login session for a user.
//=============================================================================================
func CreateSessionID(ctx context.Context, req *http.Request, userID int64) (sessionID int64, _ error) {
	agent := user_agent.New(req.Header.Get("user-agent"))
	browse, vers := agent.Browser()
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		ip = req.RemoteAddr
	}
	country := req.Header.Get("X-AppEngine-Country")
	region := req.Header.Get("X-AppEngine-Region")
	city := req.Header.Get("X-AppEngine-City")
	location, err := getLocationName(country, strings.ToUpper(region))
	if err != nil {
		location = "Unknown"
	} else {
		location = strings.Title(city) + ", " + location
	}
	newSession := Session{
		UserID:      userID,
		BrowserUsed: browse + " " + vers,
		IP:          ip,
		// Test While Deployed
		LocationUsed: location,
		LastUsed:     time.Now(),
	}
	rk, err := retrievable.PlaceEntity(ctx, int64(0), &newSession)
	if err != nil {
		return int64(-1), err
	}
	return rk.IntID(), err
}

func getLocationName(country, region string) (string, error) {
	c, err := gountries.New().FindCountryByAlpha(country)
	if err != nil {
		return "", err
	}
	for _, r := range c.SubDivisions() {
		if r.Code == region {
			return r.Name + ", " + c.Name.BaseLang.Common, nil
		}
	}
	return c.Name.BaseLang.Common, nil
}

//=============================================================================================
func GetUserIDFromSession(ctx context.Context, sessionID int64) (userID int64, _ error) {
	sessionData, err := GetSession(ctx, sessionID)
	if err != nil {
		return 0, err
	}

	return sessionData.UserID, nil
}

//=============================================================================================
func GetSession(ctx context.Context, sessionID int64) (Session, error) {
	s := Session{}
	getErr := retrievable.GetEntity(ctx, sessionID, &s) // Get actual session from datastore
	if getErr != nil {
		return Session{}, ErrNotLoggedIn
	}

	s.LastUsed = time.Now()
	if _, err := retrievable.PlaceEntity(ctx, sessionID, &s); err != nil {
		return Session{}, err
	}

	return s, nil
}

//=============================================================================================
// Gets the session ID from an active user session.
//=============================================================================================
func GetSessionID(req *http.Request) (int64, error) {
	sessionIDStr, err := GetCookieValue(req, "session")
	if err != nil {
		return -1, ErrNotLoggedIn
	}

	id, err := strconv.ParseInt(sessionIDStr, 10, 64) // Change cookie val into key
	if err != nil {
		return -1, ErrInvalidLogin
	}

	return id, nil
}

//=============================================================================================
func CleanOldSessions(ctx context.Context) error {
	q := datastore.NewQuery(SessionTable).Filter("LastUsed <", time.Now().Add(sessionTime*time.Second)).KeysOnly()
	it := q.Run(ctx)
	for {
		k, err := it.Next(nil)
		if err == datastore.Done {
			break
		} else if err != nil {
			return err
		}
		retrievable.DeleteEntity(ctx, k)
	}
	return nil
}

//=============================================================================================
func GetUserFromSession(req *http.Request) (*User, error) {
	userID, err := GetUserIDFromRequest(req)
	if err != nil {
		return &User{}, err
	}

	ctx := appengine.NewContext(req)
	return GetUserFromID(ctx, userID)
}

//=============================================================================================
func GetUserIDFromRequest(req *http.Request) (int64, error) {
	s, err := GetSessionID(req)
	if err != nil {
		return 0, err
	}

	ctx := appengine.NewContext(req)
	userID, err := GetUserIDFromSession(ctx, s)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

//=============================================================================================
func GetUserFromID(ctx context.Context, userID int64) (*User, error) {
	u := &User{}
	getErr := retrievable.GetEntity(ctx, retrievable.IntID(userID), u)
	return u, getErr
}

//=============================================================================================
func GetEmailFromID(req *http.Request, id int64) (string, error) {
	usr := User{}
	err := retrievable.GetEntity(appengine.NewContext(req), id, &usr)
	if err != nil {
		return "", err
	}
	return usr.Email, err
}

//=============================================================================================
// Checks if the user is logged in.
// If not then it redirects to the login screen.
//=============================================================================================
func MustLogin(res http.ResponseWriter, req *http.Request) bool {
	_, err := GetUserFromSession(req)
	if err != nil {
		path := strings.Replace(req.URL.Path[1:], "%2f", "/", -1)
		http.Redirect(res, req, PATH_AUTH_Login+"?redirect="+path, http.StatusSeeOther)
		return true
	}
	return false
}

//=============================================================================================
func GetUserIDFromEmail(ctx context.Context, email string) (int64, error) {
	urID := LoginLocalAccount{}
	getErr := urID.Get(ctx, email)
	return urID.UserID, getErr
}

//=============================================================================================
func GetAllSessionsForUser(ctx context.Context, userID int64) ([]Session, error) {
	q := datastore.NewQuery(SessionTable).Filter("UserID =", userID).Order("-LastUsed").Limit(50)
	results := make([]Session, 0, 50)
	it := q.Run(ctx)
	for {
		s := Session{}
		key, err := it.Next(&s)
		if err == datastore.Done {
			break
		} else if err != nil {
			return nil, err
		}
		s.StoreKey(key)
		results = append(results, s)
	}
	return results, nil
}

//=============================================================================================
func DeleteUserSession(ctx context.Context, userID int64, sessionID int64) error {
	ses := Session{}
	err := retrievable.GetEntity(ctx, sessionID, &ses)
	if err == datastore.ErrNoSuchEntity {
		return ErrNoSession
	} else if err != nil {
		return err
	}

	if ses.UserID != userID {
		return ErrNoSession
	}

	return retrievable.DeleteEntity(ctx, ses.Key(ctx, sessionID))
}

//=============================================================================================
// Deletes all sessions attached to a user except the session passed in.
// Send session ID 0 to delete all sessions.
//=============================================================================================
func DeleteAllOtherSessionsForUser(ctx context.Context, userID, keepSessionID int64) error {
	q := datastore.NewQuery(SessionTable).Filter("UserID =", userID).KeysOnly()
	it := q.Run(ctx)
	for {
		key, err := it.Next(nil)
		if err == datastore.Done {
			break
		} else if err != nil {
			return err
		}
		if key.IntID() != keepSessionID {
			retrievable.DeleteEntity(ctx, key)
		}
	}
	return nil
}

type ElevatedPerms struct {
	userid int64
}

func (e *ElevatedPerms) Key(ctx context.Context, key interface{}) *datastore.Key {
	return datastore.NewKey(ctx, "upload", "", key.(int64), nil)
}

func ReqElevatedUserPerms(ctx context.Context, userID retrievable.IntID, res http.ResponseWriter, req *http.Request) error {
	var err error
	var Elevated ElevatedPerms

	//first check to see if permission is in memcache
	err = retrievable.GetFromMemcache(ctx, strconv.FormatInt(int64(userID), 10), &Elevated)
	if err != nil && err != memcache.ErrCacheMiss {
		log.Infof(ctx, "error getting item from memcache")
	}

	//If it is not, redirect to elevated perms
	if err == memcache.ErrCacheMiss {

		path := strings.Replace(req.URL.Path[1:], "%2f", "/", -1)
		http.Redirect(res, req, "/elevatedlogin"+"?redirect="+path, http.StatusSeeOther)
		return err
	}

	//If it is, allow user to continue
	return err

}
