package flashduty

import (
	"errors"

	"github.com/ccfos/nightingale/v6/models"
	"github.com/ccfos/nightingale/v6/pkg/ctx"

	"github.com/toolkits/pkg/logger"
)

func SyncUsersChange(ctx *ctx.Context, dbUsers []*models.User, cacheUsers map[int64]*models.User) error {
	if !ctx.IsCenter {
		return nil
	}

	appKey, err := models.ConfigsGetFlashDutyAppKey(ctx)
	if err != nil {
		return err
	}

	dbUsersHas := sliceToMap(dbUsers)

	addUsers := diffMap(dbUsersHas, cacheUsers)
	if err := fdAddUsers(appKey, addUsers); err != nil {
		return err
	}

	delUsers := diffMap(cacheUsers, dbUsersHas)
	fdDelUsers(appKey, delUsers)

	return nil
}

func sliceToMap(dbUsers []*models.User) map[int64]*models.User {
	m := make(map[int64]*models.User, len(dbUsers))
	for _, user := range dbUsers {
		m[user.Id] = user
	}
	return m
}

// in m1 and not in m2
func diffMap(m1, m2 map[int64]*models.User) []models.User {
	var diff []models.User
	for i := range m1 {
		if _, ok := m2[i]; !ok {
			diff = append(diff, *m1[i])
		}
	}
	return diff
}

type User struct {
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	CountryCode string `json:"country_code"`
	MemberName  string `json:"member_name"`
	RoleIds     []int  `json:"role_ids"`
}

func (user *User) delMember(appKey string) error {
	if user.Email == "" && user.Phone == "" {
		return errors.New("phones and email must be selected one of two")
	}
	return PostFlashDuty("/member/delete", appKey, user)
}

type Members struct {
	Users []User `json:"members"`
}

func (m *Members) addMembers(appKey string) error {
	if len(m.Users) == 0 {
		return nil
	}
	validUsers := make([]User, 0, len(m.Users))
	for _, user := range m.Users {
		if user.Email == "" && (user.Phone == "" || user.MemberName == "") {
			logger.Errorf("user(%v) phone and email must be selected one of two, and the member_name must be added when selecting the phone", user)
		} else {
			validUsers = append(validUsers, user)
		}
	}
	m.Users = validUsers
	return PostFlashDuty("/member/invite", appKey, m)
}

func fdAddUsers(appKey string, users []models.User) error {
	fdUsers := usersToFdUsers(users)
	members := &Members{
		Users: fdUsers,
	}
	return members.addMembers(appKey)
}

func fdDelUsers(appKey string, users []models.User) {
	fdUsers := usersToFdUsers(users)
	for _, fdUser := range fdUsers {
		if err := fdUser.delMember(appKey); err != nil {
			logger.Error("failed to delete user: %v", err)
		}
	}
}

func usersToFdUsers(users []models.User) []User {
	fdUsers := make([]User, 0, len(users))
	for i := range users {
		fdUsers = append(fdUsers, User{
			Email:      users[i].Email,
			Phone:      users[i].Phone,
			MemberName: users[i].Username,
		})
	}
	return fdUsers
}
