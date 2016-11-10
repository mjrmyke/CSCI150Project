// Important global variables
package main

const (
	gcsBucket  = "http://csci150project.appspot.com/"
	DomainPath = "http://localhost:8080/"
	// TODO Change this to a more random key value
	HMAC_Key = "csci150project2016"
)

// This key needs to be exactly 32 bytes long
// TODO This should not be in our git repo
var encryptKey = []byte{33, 44, 160, 6, 124, 138, 93, 47, 177, 135, 163, 154, 42, 14, 58, 17, 85, 133, 174, 207, 255, 52, 3, 26, 145, 21, 169, 65, 106, 108, 0, 66}
