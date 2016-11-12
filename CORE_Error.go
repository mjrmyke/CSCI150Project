package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

// Error definitions here.
var (
	ErrPasswordMatch           = errors.New("Passwords fields do not match")
	ErrInvalidLogin            = errors.New("Error Login: Parameters may not be nil!")
	ErrNotLoggedIn             = errors.New("Login Error: Cannot verify session!")
	ErrUsernameExists          = errors.New("Validation Error: Username already exists!")
	ErrInvalidPermission       = errors.New("Permission Error: Not Allowed")
	ErrNotImplemented          = errors.New("Structure Error: Function Not Implemented!")
	ErrEmptyField              = errors.New("Input Error: Required Field Empty")
	ErrNoUser                  = errors.New("Login Error: No Such User")
	ErrNoSession               = errors.New("Session does not exist")
	ErrTooLarge                = errors.New("Image Dimensions too large")
	ErrInvalidEmail            = errors.New("The email sent is invalid")
	ErrCategoryDoesNotExist    = errors.New("Category does not exist")
	ErrSubCategoryDoesNotExist = errors.New("Sub-Category does not exist")
	ErrMustOwnNotes            = errors.New("User does not own note")
	ErrMustOwnUpload           = errors.New("User does not own upload")
	ErrMustOwnItem             = errors.New("User does not own item")
	ErrNotMatchingHMac         = errors.New("Hmac checking failed")
	ErrNoItemToDelete          = errors.New("Must specify item to delete")
	ErrItemDoesNotExist        = errors.New("Item does not exist")
)

// Error based functions here.

// Internal Function: ErrorPage
/// Prints an error page to response and returns a boolean representation of the function executing.
/// Results: Boolean Value
////  True: Parent should cease execution, error has been found.
////  False: No Error, Parent may ignore this function.
/// Usage: Use if there is no constructive alternative.
func ErrorPage(ctx context.Context, res http.ResponseWriter, u *User, ErrorTitle string, e error, errCode int) bool {
	if e != nil {
		log.Errorf(ctx, "%s ---- %v\n", ErrorTitle, e)
		if u == nil {
			u = &User{}
		}
		args := &struct {
			Header    HeaderData
			ErrorName string
			ErrorDump error
			ErrorCode int
		}{
			HeaderData{true, true, ctx, u, ""}, ErrorTitle, e, errCode,
		}
		res.WriteHeader(errCode)
		ServeTemplateWithParams(res, "site-error", args) // Execute the error page with the anonymous struct.
		return true
	}
	return false
}

// Internal Function: BackWithError
/// Returns to GET responding with FormValue("ErrorResponse")
/// Results: Boolean Value
////  True: Parent should cease execution, error has been found.
////  False: No Error, Parent may ignore this function.
/// Usage: Use in POST calls accessed from a GET of the same handle.
func BackWithError(res http.ResponseWriter, req *http.Request, err error, errorString string) bool {
	if err != nil {
		path := strings.Replace(req.URL.Path, "%2f", "/", -1)
		http.Redirect(res, req, path+"?"+url.QueryEscape("ErrorResponse")+"="+url.QueryEscape(errorString), http.StatusSeeOther)
		return true
	}
	return false
}

// PanicResponse:

// PanicResponse:
// A local struct used within this package for the ThrowPanic function.
type PanicResponse struct {
	caller string
	code   int
	error
}

// Only use this in the case of something that shows that the site is broken.
func ThrowPanic(caller string, code int, e error) {
	if e == nil {
		panic(PanicResponse{caller, code, errors.New("")})
	}
	panic(PanicResponse{caller, code, e})
}
func PanicPage(res http.ResponseWriter, req *http.Request, param interface{}) {
	fmt.Fprint(res, `<html><plaintext>`)
	if pr, b := param.(PanicResponse); b {
		fmt.Fprintln(res, "An Error has occurred.")
		fmt.Fprintln(res, "Caller:", pr.caller)
		fmt.Fprintln(res, "Code:", pr.code)
		fmt.Fprintln(res, "Details:", pr.Error())
	} else {
		fmt.Fprintln(res, "Error:", param)
	}
}

func DestinationWithError(res http.ResponseWriter, req *http.Request, src string, err error, errorString string) bool {
	return DestinationWithErrorAt(res, req, src, err, "ErrorResponse", errorString)
}

func DestinationWithErrorAt(res http.ResponseWriter, req *http.Request, src string, err error, errorLocation, errorString string) bool {
	if err != nil {
		path, parseErr := url.Parse(src)
		if parseErr != nil {
			ThrowPanic("DestinationWithErrorAt", http.StatusInternalServerError, parseErr)
		}
		path.RawQuery = url.QueryEscape(errorLocation) + "=" + url.QueryEscape(errorString)
		http.Redirect(res, req, path.String(), http.StatusSeeOther)
		return true
	}
	return false
}
