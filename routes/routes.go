package routes

import (
	"we-credit/controllers"
	"we-credit/docs"
	"we-credit/utility"

	"github.com/gin-gonic/gin"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// AddRoutes is responsible for adding all the routes so the server can handle
// new routes. this means that we can reuse this function for multiple prefixes.
// prefixes like job_portal are necessary for legacy url handling.
func AddRoutes(router *gin.RouterGroup) {

	// NOTE :- all api must be in this group and for every particular feature apis must be create new group.
	api := router.Group("/user")
	{

		// This api is responsible for student registration
		api.POST("/authenticate", controllers.UserRegistration)
		api.POST("/otp/verify", controllers.VerifyCode)
		// This api is responsible for resend otp on phone number.
		router.POST("/otp/send", controllers.ResendVerificationCode)

	}

}

// SetupRouter sets up routes
func SetupRouter() *gin.Engine {

	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/user"
	url := ginSwagger.URL(utility.GetHostURL() + "/swagger/doc.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, url))
	// Add all current URls
	AddRoutes(&router.RouterGroup)
	return router
}
