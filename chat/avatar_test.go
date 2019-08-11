package main

import "testing"

func TestAuthAvatar(t *testing.T) {
	var authAvatar AuthAvatar
	client := new(client)

	url, err := authAvatar.GetAvatarURL(client)
	if err != ErrNoAvatarURL {
		t.Error("AuthAvatar.GetAvatarURL should return ErrNoAvatarURL when no value present")
	}
	// set a value
	testUrl := "https://url-to-gravatar/"
	client.userData = &User{AvatarURL: testUrl}
	url, err = authAvatar.GetAvatarURL(client)
	if err != nil {
		t.Error("AuthAvatar.GetAvatarURL should return no error when value is present")
	}
	if url != testUrl {
		t.Errorf("AuthAvatar.GetAvatarURL should return %s. got %s", testUrl, url)
	}
}
