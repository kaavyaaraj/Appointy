package instag

import (
	"encoding/json"
	"fmt"
)

type accountResp struct {
	Status  string  `json:"status"`
	Account Account `json:"logged_in_user"`
}
type Account struct {
	inst *Instagram

	ID                         int64        `json:"pk"`
	Username                   string       `json:"username"`
	FullName                   string       `json:"full_name"`
	Biography                  string       `json:"biography"`
	ProfilePicURL              string       `json:"profile_pic_url"`
	Email                      string       `json:"email"`
	PhoneNumber                string       `json:"phone_number"`
	
}

func (account *Account) Sync() error {
	insta := account.inst
	data, err := insta.prepareData()
	if err != nil {
		return err
	}
	body, err := insta.sendRequest(&reqOptions{
		Endpoint: urlCurrentUser,
		Query:    generateSignature(data),
		IsPost:   true,
	})
	if err == nil {
		resp := profResp{}
		err = json.Unmarshal(body, &resp)
		if err == nil {
			*account = resp.Account
			account.inst = insta
		}
	}
	return err
}

// ChangePassword changes current password.
func (account *Account) ChangePassword(old, new string) error {
	insta := account.inst
	data, err := insta.prepareData(
		map[string]interface{}{
			"old_password":  old,
			"new_password1": new,
			"new_password2": new,
		},
	)
	if err == nil {
		_, err = insta.sendRequest(
			&reqOptions{
				Endpoint: urlChangePass,
				Query:    generateSignature(data),
				IsPost:   true,
			},
		)
	}
	return err
}

type profResp struct {
	Status  string  `json:"status"`
	Account Account `json:"user"`
}

func (account *Account) Followers() *Users {
	endpoint := fmt.Sprintf(urlFollowers, account.ID)
	users := &Users{}
	users.inst = account.inst
	users.endpoint = endpoint
	return users
}

func (account *Account) Following() *Users {
	endpoint := fmt.Sprintf(urlFollowing, account.ID)
	users := &Users{}
	users.inst = account.inst
	users.endpoint = endpoint
	return users
}

// For pagination use FeedMedia.Next()
func (account *Account) Archived(params ...interface{}) *FeedMedia {
	insta := account.inst

	media := &FeedMedia{}
	media.inst = insta
	media.endpoint = urlUserArchived

	for _, param := range params {
		switch s := param.(type) {
		case string:
			media.timestamp = s
		}
	}

	return media
}
