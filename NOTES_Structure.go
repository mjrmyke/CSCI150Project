package main

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// TODO: utilize
type Note struct {
	OwnerID int64			// owner of the note, can set permissions
	//Collab []int64			// any people collabing, stretch goal
	Protected bool			// if true then it is not publically editable.
	ContentID int64			// Reference to the content of the note.
}
type Content struct{
	Title,Content string	// Content can be raw html or markdown
}

// int64 Keys
func (n *Note) Key(ctx context.Context, key interface{}) *datastore.Key {
	return datastore.NewKey(ctx, "NotePermissions", "", key.(int64), nil)
}
func (n *Content) Key(ctx context.Context, key interface{}) *datastore.Key {
	return datastore.NewKey(ctx, "NotePermissions", "", key.(int64), nil)
}