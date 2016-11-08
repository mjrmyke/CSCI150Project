// Contains data structures needed for our OAuth interactions.
package main
import(
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"	
)
const ( 
	OauthTable = "Oauth" 
)
type LoginOauthAccount struct {
	UserID int64
}
// String Keys
func (l *LoginOauthAccount) Key(ctx context.Context, key interface{}) *datastore.Key {
	return datastore.NewKey(ctx, OauthTable, key.(string), 0, nil)
}