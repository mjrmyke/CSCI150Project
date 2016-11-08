package main

import (
	"strconv"

	"github.com/Esseh/retrievable"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

//if a user has permission to view a note,
//There will exist an entry in the db with the id of
//NoteID concatenated to UserID
//NOTEIDUSERID
type NotePermissions struct {
	OwnerId int64
}

// String Keys
func (n *NotePermissions) Key(ctx context.Context, key interface{}) *datastore.Key {
	return datastore.NewKey(ctx, "NotePermissions", key.(string), 0, nil)
}

//function to add users permission for Notes
func AddNotePerms(ctx context.Context, owner, useridtoadd int64) {
	owneridstr := strconv.FormatInt(owner, 10)
	useridtoaddstr := strconv.FormatInt(useridtoadd, 10)
	concatstr := owneridstr + useridtoaddstr
	NoteStruct := &NotePermissions{
		OwnerId: owner,
	}
	retrievable.PlaceInDatastore(ctx, concatstr, NoteStruct)
}

//function to remove users permission for Notes
func RemoveNotePerms(ctx context.Context, owner, useridtoremove int64) error {
	owneridstr := strconv.FormatInt(owner, 10)
	useridtoremovestr := strconv.FormatInt(useridtoremove, 10)
	concatstr := owneridstr + useridtoremovestr
	err := retrievable.DeleteEntity(ctx, (&NotePermissions{}).Key(ctx, concatstr))
	return err
}
