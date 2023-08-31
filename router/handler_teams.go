package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"strconv"
	"time"
)

func (srv *Service) createTeam(ctx *gin.Context) {
	var team Team
	if err := ctx.ShouldBindJSON(&team); err != nil {
		ReturnErrorWithMsg(ctx, InvalidFormData, err.Error())
		return
	}
	var t Team
	if err := srv.mysql.Client.Model(Team{}).Where("handle = ?", team.Handle).First(&t).Error; err == nil {
		ReturnErrorWithMsg(ctx, InternalError, "handle already exist")
		return
	}
	browserUser, _ := srv.getBrowserUserFromContext(ctx)
	team.UserId = browserUser.UserId

	srv.mysql.Client.Create(&team)
	var record = TeamUser{
		TeamId: team.Id,
		UserId: browserUser.UserId,
		Role:   RoleAdmin,
		Status: TeamUserStatusJoined,
	}

	srv.mysql.Client.Create(&record)
	ReturnSuccess(ctx, struct{}{})
	return
}

func (srv *Service) updateTeam(ctx *gin.Context) {
	tId := ctx.Param("team_id")
	var team Team
	if err := ctx.ShouldBindJSON(&team); err != nil {
		ReturnErrorWithMsg(ctx, InvalidFormData, err.Error())
		return
	}
	srv.mysql.Client.Model(team).Where("id = ?", tId).Omit("user_id", "handle").Updates(&team)

	ReturnSuccess(ctx, struct{}{})
	return
}

func (srv *Service) deleteTeam(ctx *gin.Context) {
	var team Team
	if err := ctx.ShouldBindJSON(&team); err != nil {
		ReturnErrorWithMsg(ctx, InvalidFormData, err.Error())
		return
	}

	srv.mysql.Client.Where("id = ?", team.Id).Omit("user_id").Create(&team)

	ReturnSuccess(ctx, struct{}{})
	return
}

func (srv *Service) getTeams(ctx *gin.Context) {
	var records []Team
	var filters Filters
	if name := ctx.Query("name"); name != "" {
		filters = filters.Add("name like ?", "%"+name+"%")
	}

	browserUser, _ := srv.getBrowserUserFromContext(ctx)
	filters = filters.Add("user_id = ?", browserUser.UserId)
	order := NewOrderBy()

	conditions := DbConditions{
		Model:   &records,
		Filters: &filters,
		Order:   &order,
	}
	data := GetRecords(srv.mysql.Client, &conditions)

	ReturnSuccess(ctx, data)
	return
}

func (srv *Service) getTeam(ctx *gin.Context) {
	id := ctx.Param("team_id")

	var record Team
	var filters Filters
	filters = filters.Add("id = ?", id)

	conditions := DbConditions{
		Model:   &record,
		Filters: &filters,
	}

	_, _ = GetRecord(srv.mysql.Client, &conditions)
	ReturnSuccess(ctx, record)
	return
}

func (srv *Service) getTeamMember(ctx *gin.Context) {
	teamId := ctx.Param("team_id")
	var records []TeamUser
	var filters Filters

	filters = filters.Add("team_id = ?", teamId)
	order := NewOrderBy()

	conditions := DbConditions{
		Model:    &records,
		Filters:  &filters,
		Order:    &order,
		Preloads: []string{"Team", "User"},
	}
	_ = GetRecords(srv.mysql.Client, &conditions)
	for i, user := range records {
		if user.InviteEmail != "" {
			records[i].User.UserName = ""
			records[i].User.Email = user.InviteEmail
			records[i].User.Email = user.InviteEmail
		}
		if user.Status == TeamUserStatusPending && time.Now().Unix() > user.ExpireAt {
			records[i].Status = TeamUserStatusExpired
		}
	}

	ReturnSuccess(ctx, records)
	return
}

func (srv *Service) addTeamMember(ctx *gin.Context) {
	var body ParamAddTeamMember
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ReturnErrorWithMsg(ctx, InvalidFormData, err.Error())
		return
	}
	_, err := srv.getTeamById(body.TeamId)
	if err != nil {
		ReturnErrorWithMsg(ctx, InvalidFormData, "team not exist")
		return
	}
	if body.Role != RoleAdmin && body.Role != RoleOperator && body.Role != RoleViewer {
		ReturnErrorWithMsg(ctx, InvalidFormData, fmt.Sprintf("wrong role, only %s, %s, %s", RoleAdmin, RoleOperator, RoleViewer))
		return
	}
	var inviteCode string
	user, _ := srv.getUserByEmail(body.Email)

	inviteCode = srv.generateInviteCode(&TeamUser{})
	go srv.sendEmail(ctx, body.Email, inviteCode)

	teamUser, err := srv.getTeamUser(body.TeamId, user.Id)
	if err == nil {
		if teamUser.Status == TeamUserStatusPending {
			ReturnErrorWithMsg(ctx, InvalidFormData, "Waiting for an invitation reply")
			return
		} else {
			ReturnErrorWithMsg(ctx, InvalidFormData, "The user is already on the team")
			return
		}
	}
	var record TeamUser
	record.TeamId = body.TeamId
	record.UserId = user.Id
	record.Status = TeamUserStatusPending
	record.Role = body.Role
	record.InviteEmail = body.Email
	record.InviteCode = inviteCode
	record.ExpireAt = time.Now().Unix() + 86400
	srv.mysql.Client.Create(&record)
	ReturnSuccess(ctx, struct{}{})
	return
}

func (srv *Service) delTeamMember(ctx *gin.Context) {
	memId := ctx.Param("mem_id")

	var filters Filters
	filters = filters.Add("id = ?", memId)
	var record TeamUser
	_, _ = DeleteRecord(srv.mysql.Client, &record, &filters)

	data := struct{}{}
	ReturnSuccess(ctx, data)
}

func (srv *Service) getMyTeams(ctx *gin.Context) {
	var records []TeamUser
	var filters Filters
	paginatorParam := GetNewPaginator()

	browserUser, _ := srv.getBrowserUserFromContext(ctx)
	filters = filters.Add("user_id = ?", browserUser.UserId)
	filters = filters.Add("status = ?", TeamUserStatusJoined)
	order := NewOrderBy()

	conditions := DbConditions{
		Fields:         []string{"team_id", "role", "created_at"},
		Model:          &records,
		Filters:        &filters,
		Order:          &order,
		Preloads:       []string{"Team"},
		PaginatorParam: paginatorParam,
	}
	data := GetRecords(srv.mysql.Client, &conditions)

	ReturnSuccess(ctx, data)
	return
}

func (srv *Service) dealInvitations(ctx *gin.Context) {
	id := ctx.Param("id")
	resp := ctx.Param("response")

	inviteId, _ := strconv.Atoi(id)
	invitation, err := srv.getTeamUserById(int64(inviteId))
	if err != nil {
		ReturnErrorWithMsg(ctx, InternalError, "invitation not exist")
		return
	}
	browserUser, _ := srv.getBrowserUserFromContext(ctx)
	if invitation.UserId != browserUser.UserId {
		ReturnErrorWithMsg(ctx, InternalError, "Unauthorized operation")
		return
	}

	if resp != InvitationAccept && resp != InvitationReject {
		ReturnErrorWithMsg(ctx, InternalError, "only Accept or Reject")
		return
	}
	var status TeamUserStatus
	if resp == "Accept" {
		status = TeamUserStatusJoined
	} else {
		status = TeamUserStatusReject
	}
	data := make(map[string]interface{})
	data["status"] = status
	srv.mysql.Client.Model(TeamUser{}).Where("id = ?", id).Updates(data)
	if status == TeamUserStatusReject {
		srv.mysql.Client.Where("id = ?", id).Delete(TeamUser{})
	}
	ReturnSuccess(ctx, struct{}{})
	return
}

func (srv *Service) updateTeamMember(ctx *gin.Context) {
	memId := ctx.Param("mem_id")
	var r ParamUpdateTeamMember
	if err := ctx.ShouldBindWith(&r, binding.JSON); err != nil {
		ReturnErrorWithMsg(ctx, InvalidFormData, err.Error())
		return
	}
	if r.Role != RoleAdmin && r.Role != RoleOperator && r.Role != RoleViewer {
		ReturnErrorWithMsg(ctx, InvalidFormData, fmt.Sprintf("wrong role, only %s, %s, %s", RoleAdmin, RoleOperator, RoleViewer))
		return
	}

	data := make(map[string]interface{})
	data["role"] = r.Role
	err := srv.mysql.Client.Model(TeamUser{}).Where("id = ?", memId).Updates(data).Error

	if err != nil {
		ReturnErrorWithMsg(ctx, InternalError, "Operation failed")
		return
	}
	ReturnSuccess(ctx, struct{}{})
}
