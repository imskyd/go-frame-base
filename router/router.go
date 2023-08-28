package router

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/raven-ruiwen/go-helper/auth0"
)

func (srv *Service) GetRouter() *gin.Engine {
	router := gin.Default()
	router.Use(gzip.Gzip(gzip.BestSpeed))
	router = auth0.RegisterRouter(router, []byte("secret"))

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("usernamevalidator", usernameValidator)
	}

	v1Public := router.Group(srv.prefix)
	{
		//login
		v1Public.GET("/login", srv.login)
		v1Public.GET("/team/invite", srv.teamInvite)
		v1Public.GET("/login/local", srv.loginLocal)
		v1Public.GET("/logout", srv.logout)
		v1Public.GET("/login/callback", srv.auth0CallBack)
	}

	//middleware, after this, all api need login
	router.Use(srv.userMustLoginMiddleWare())

	v1 := router.Group(srv.prefix)
	{
		v1.GET("/mine/profile", srv.getProfile)
		v1.PUT("/mine/profile", srv.updateProfile)
		v1.Use(srv.userStatusMiddleWare())
		v1.GET("/mine/invitations", srv.getInvitations)
		v1.DELETE("/user", srv.userDelete)

		teams := v1.Group("/teams")
		{
			teams.POST("", srv.createTeam)
			teams.Use(srv.teamBasicMiddleWare())
			teams.GET("", srv.getMyTeams)
			teams.GET("/:team_id", srv.getTeam)
			teams.GET("/:team_id/members", srv.getTeamMember)

			teams.Use(srv.teamOperateMiddleWare())
			teams.DELETE("/:team_id", srv.deleteTeam)
			teams.PUT("/:team_id", srv.updateTeam)
			teams.POST("/:team_id/members", srv.addTeamMember)
			teams.PUT("/:team_id/members/:mem_id", srv.updateTeamMember)
			teams.DELETE("/:team_id/members/:mem_id", srv.delTeamMember)
		}
	}

	return router
}
