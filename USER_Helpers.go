package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"strconv"

	"golang.org/x/net/context"
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
