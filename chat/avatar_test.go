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

func TestGravatarAvatar(t *testing.T) {
	var gravatarAvatar GravatarAvatar
	client := new(client)
	client.userData = &User{userId: "0bc83cb571cd1c50ba6f3e8a78ef1346"}
	url, err := gravatarAvatar.GetAvatarURL(client)
	if err != nil {
		t.Error("GravatarAvatar.GetAvatarURL should not return an error")
	}
	if url != "//www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346" {
		t.Errorf("GravatarAvatar.GetAvatarURL should return %s. got %s", "//www.gravatar.com/avatar/0bc83cb571cd1c50ba6f3e8a78ef1346", url)
	}
}
