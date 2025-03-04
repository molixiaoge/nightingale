package flashduty

import (
	"errors"

	"github.com/ccfos/nightingale/v6/models"
	"github.com/ccfos/nightingale/v6/pkg/ctx"

	"github.com/toolkits/pkg/logger"
)

type UserGroupSyncer struct {
	ctx    *ctx.Context
	ug     *models.UserGroup
	appKey string
}

func NewUserGroupSyncer(ctx *ctx.Context, ug *models.UserGroup) (*UserGroupSyncer, error) {
	appKey, err := models.ConfigsGetFlashDutyAppKey(ctx)
	if err != nil {
		return nil, err
	}

	return &UserGroupSyncer{
		ctx:    ctx,
		ug:     ug,
		appKey: appKey,
	}, nil
}

func (ugs *UserGroupSyncer) SyncUGAdd() error {
	return ugs.syncTeamMember()
}

func (ugs *UserGroupSyncer) SyncUGPut(oldUGName string) error {
	if err := ugs.syncTeamMember(); err != nil {
		return err
	}

	if oldUGName != ugs.ug.Name {
		if err := ugs.SyncUGDel(oldUGName); err != nil {
			return err
		}
	}
	return nil
}

func (ugs *UserGroupSyncer) SyncUGDel(ugName string) error {
	fdt := Team{
		TeamName: ugName,
	}
	err := fdt.DelTeam(ugs.appKey)
	return err
}

func (ugs *UserGroupSyncer) SyncMembersAdd() error {
	return ugs.syncTeamMember()
}

func (ugs *UserGroupSyncer) SyncMembersDel() error {
	return ugs.syncTeamMember()
}

func (ugs *UserGroupSyncer) syncTeamMember() error {
	uids, err := models.MemberIds(ugs.ctx, ugs.ug.Id)
	if err != nil {
		return err
	}
	users, err := models.UserGetsByIds(ugs.ctx, uids)
	if err != nil {
		return err
	}
	err = ugs.addMemberToFDTeam(users)

	return err
}

func (ugs *UserGroupSyncer) addMemberToFDTeam(users []models.User) error {
	if err := fdAddUsers(ugs.appKey, users); err != nil {
		return err
	}

	emails := make([]string, 0)
	phones := make([]string, 0)
	for _, user := range users {
		if user.Email != "" {
			emails = append(emails, user.Email)
		} else if user.Phone != "" {
			phones = append(phones, user.Phone)
		} else {
			logger.Warningf("The user %s has no email and phone, and failed to sync to flashduty's team", user.Username)
		}
	}

	fdt := Team{
		TeamName: ugs.ug.Name,
		Emails:   emails,
		Phones:   phones,
	}
	err := fdt.UpdateTeam(ugs.appKey)
	return err
}

type Team struct {
	TeamName         string   `json:"team_name"`
	ResetIfNameExist bool     `json:"reset_if_name_exist"`
	Description      string   `json:"description"`
	Emails           []string `json:"emails"`
	Phones           []string `json:"phones"`
}

func (t *Team) AddTeam(appKey string) error {
	if t.TeamName == "" {
		return errors.New("team_name must be set")
	}
	return PostFlashDuty("/team/upsert", appKey, t)
}

func (t *Team) UpdateTeam(appKey string) error {
	t.ResetIfNameExist = true
	err := t.AddTeam(appKey)
	return err
}

func (t *Team) DelTeam(appKey string) error {
	if t.TeamName == "" {
		return errors.New("team_name must be set")
	}
	return PostFlashDuty("/team/delete", appKey, t)
}
