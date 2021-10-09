
package insta

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Users struct {
	inst *Instagram

	err      error
	endpoint string

	Status    string          `json:"status"`
	BigList   bool            `json:"big_list"`
	Users     []User          `json:"users"`
	PageSize  int             `json:"page_size"`
	RawNextID json.RawMessage `json:"next_max_id"`
	NextID    string          `json:"-"`
}

func newUsers(inst *Instagram) *Users {
	users := &Users{inst: inst}

	return users
}

func (users *Users) SetInstagram(inst *Instagram) {
	users.inst = inst
}

var ErrNoMore = errors.New("List end have been reached")

func (users *Users) Next() bool {
	if users.err != nil {
		return false
	}

	insta := users.inst
	endpoint := users.endpoint

	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint: endpoint,
			Query: map[string]string{
				"max_id":             users.NextID,
				"ig_sig_key_version": goInstaSigKeyVersion,
				"rank_token":         insta.rankToken,
			},
		},
	)
	if err == nil {
		usrs := Users{}
		err = json.Unmarshal(body, &usrs)
		if err == nil {
			if len(usrs.RawNextID) > 0 && usrs.RawNextID[0] == '"' && usrs.RawNextID[len(usrs.RawNextID)-1] == '"' {
				if err := json.Unmarshal(usrs.RawNextID, &usrs.NextID); err != nil {
					users.err = err
					return false
				}
			} else {
				var nextID int64
				if err := json.Unmarshal(usrs.RawNextID, &nextID); err != nil {
					users.err = err
					return false
				}
				usrs.NextID = strconv.FormatInt(nextID, 10)
			}
			*users = usrs
			if !usrs.BigList || usrs.NextID == "" {
				users.err = ErrNoMore
			}
			users.inst = insta
			users.endpoint = endpoint
			users.setValues()
			return true
		}
	}
	users.err = err
	return false
}

func (users *Users) Error() error {
	return users.err
}

func (users *Users) setValues() {
	for i := range users.Users {
		users.Users[i].inst = users.inst
	}
}

type userResp struct {
	Status string `json:"status"`
	User   User   `json:"user"`
}

type User struct {
	inst *Instagram

	ID                         int64   `json:"pk"`
	Username                   string  `json:"username"`
	FullName                   string  `json:"full_name"`
	Biography                  string  `json:"biography"`
	ProfilePicURL              string  `json:"profile_pic_url"`
	Email                      string  `json:"email"`
	PhoneNumber                string  `json:"phone_number"`
	
}

func (user *User) SetInstagram(insta *Instagram) {
	user.inst = insta
}


func (inst *Instagram) NewUser() *User {
	return &User{inst: inst}
}


func (user *User) Sync(params ...interface{}) error {
	insta := user.inst
	body, err := insta.sendSimpleRequest(urlUserInfo, user.ID)
	if err == nil {
		resp := userResp{}
		err = json.Unmarshal(body, &resp)
		if err == nil {
			*user = resp.User
			user.inst = insta
			for _, param := range params {
				switch b := param.(type) {
				case bool:
					if b {
						err = user.FriendShip()
					}
				}
			}
		}
	}
	return err
}


// Users.Next can be used to paginate

func (user *User) Following() *Users {
	users := &Users{}
	users.inst = user.inst
	users.endpoint = fmt.Sprintf(urlFollowing, user.ID)
	return users
}




func (user *User) Feed(params ...interface{}) *FeedMedia {
	insta := user.inst

	media := &FeedMedia{}
	media.inst = insta
	media.endpoint = urlUserFeed
	media.uid = user.ID

	for _, param := range params {
		switch s := param.(type) {
		case string:
			media.timestamp = s
		}
	}

	return media
}

