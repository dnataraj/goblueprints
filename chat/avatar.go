package main

import "errors"

// ErrNoAvatarURL is the error that is returned when the Avatar instance
// is unable to provide avatar URL
var ErrNoAvatarURL = errors.New("chat: unable to get an avatar URL")

// Avatar represents types capable of representing user profile pictures
type Avatar interface {
	// GetAvatarURL gets the avatar URL for the specified client,
	// or returns an error if something goes wrong.
	// Returns ErrNoAvatarURL if the object is unable to get an URL
	// for the specified client.
	GetAvatarURL(c *client) (string, error)
}

type AuthAvatar struct{}

var UseAuthAvatar AuthAvatar

func (AuthAvatar) GetAvatarURL(c *client) (string, error) {
	if c.userData == nil || len(c.userData.AvatarURL) == 0 {
		return "", ErrNoAvatarURL
	}
	return c.userData.AvatarURL, nil
}