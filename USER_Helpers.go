package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"strconv"

	"github.com/Esseh/retrievable"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const maxAvatarSize = 500

func getAvatarPath(userID int64) string {
	return "users/" + strconv.FormatInt(userID, 10) + "/avatar"
}

func UploadAvatar(ctx context.Context, userID int64, header *multipart.FileHeader, avatarReader io.ReadSeeker) error {
	m, _, err := image.Decode(avatarReader)
	if err != nil {
		return err
	}
	imageBounds := m.Bounds()
	if imageBounds.Dy() > maxAvatarSize || imageBounds.Dx() > maxAvatarSize {
		return ErrTooLarge
	}
	if _, err = avatarReader.Seek(0, 0); err != nil {
		return err
	}
	filename := getAvatarPath(userID)
	return AddFileToGCS(ctx, filename, header.Header["Content-Type"][0], avatarReader)
}

type NoteOutput struct {
	ID      int64
	Data    Note
	Content Content
}

// Gets all the notes made my the user.
func GetAllNotes(ctx context.Context, userID int64) ([]NoteOutput, error) {
	q := datastore.NewQuery(NoteTable).Filter("OwnerID =", userID)
	res := []Note{}
	output := make([]NoteOutput, 0)
	keys, err := q.GetAll(ctx, &res)
	if err != nil {
		return nil, err
	}
	for i, _ := range res {
		var c Content
		err := retrievable.GetEntity(ctx, res[i].ContentID, &c)
		if err != nil {
			return nil, err
		}
		output = append(output, NoteOutput{keys[i].IntID(), res[i], c})
	}
	return output, nil
}
