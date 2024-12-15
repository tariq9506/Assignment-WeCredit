package controllers

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"we-credit/models"
	"we-credit/service"
	"we-credit/utility"

	"github.com/gin-gonic/gin"
)

// UserRegistration godoc
// @Summary This controller will handles the registration process for user.
// @description The UserRegistration function handles the process of signing up
// @description user on a website. It expects the user to submit their phone number through a form.
// @Tags Registration
// @Accept application/x-www-form-urlencoded
// @Param phone-number  formData  string true "Phone"
// @Produce json
// @Success 200
// @Router /authenticate [post]
func UserRegistration(c *gin.Context) {
	user := RegisterUser(c)
	// All errors are handled in the 'RegisterUser' function.
	// If any error occurs, 'RegisterUser' returns an error in c.JSON and an empty 'User' struct.
	// We check if the struct is empty then return from the function to prevent sending both an error and a success response.
	if user.ID == 0 {
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success",
		"message":        "One time message has been sent to you phone number.",
		"user_id":        user.ID,
		"phone_verified": user.IsPhoneVerified,
	})
}
func RegisterUser(c *gin.Context) models.User {

	userIP := utility.GetClientIP(c)
	phoneNumber := c.PostForm("phone-number")
	// Check if the phone number is empty.
	if len(phoneNumber) == 0 {
		log.Println("Registration Failed: Phone number not found. Please enter a valid phone number.")
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Failed",
			"message": "Please enter a valid number",
		})
		return models.User{}
	}
	// Use a regular expression to validate the phone number format (exactly 10 digits).
	match, err := regexp.MatchString("^[0-9]{10}$", phoneNumber)
	if !match || err != nil {
		log.Println("RegisterUser: Failed, invalid phone number with error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Failed",
			"message": "Please enter a valid number",
		})
		return models.User{}
	}
	location := service.GetLocationFromIP(userIP)
	details, err := models.GetDetailsOfSupportedCountryByCode(location.CountryCode)
	if err != nil {
		log.Println("RegisterUser: GetDetailsOfSupportedCountryByCode failed to get location information with error: ", err)
	}
	otp := utility.GenerateOTP()
	user := models.User{
		Phone:       phoneNumber,
		DialingCode: details.CountryPhoneCode,
		UserIP:      userIP,
		Location:    location,
		OTP:         otp,
	}
	// This func will check is tht given phone number deliverable or not, if not deliverable will return an error message
	isDeliverable, err := service.IsPhNumberDeliverable(phoneNumber, details.CountryCode)
	if !isDeliverable {
		log.Println("RegisterUser: failed phone number lookup with flag :", isDeliverable)
		c.JSON(http.StatusOK, gin.H{
			"status":  "Failed",
			"message": "Please enter a valid number",
		})
		return models.User{}
	}
	// This func will check is tht given phone number voip or not, if not deliverable will return an error message
	isVoipNumberAllowed := utility.AllowVoipNumbers()
	if !isVoipNumberAllowed {
		isVoip, err := service.IsPhoneNumberVoip(phoneNumber, details.CountryCode)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status":  "Failed",
				"message": "Please enter a valid number",
				"error":   err,
			})
		}
		if isVoip {
			log.Println("RegisterUser: failed while checking phone number is voip :", isVoip)
			c.JSON(http.StatusOK, gin.H{
				"status":  "Failed",
				"message": "Please enter a valid number",
			})
			return models.User{}
		}
	}

	// func to send the verification code to the user's phone number
	err = SendPhoneNumberVerificationCode(user)
	if err != nil {
		log.Println("Registration Failed: Unable to send verification code. Please try again later", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Failed",
			"message": "Failed to send otp.",
		})
		return models.User{}
	}

	_, _ = models.SaveNewUser(&user)
	return user
}

// GetStudentProfile godoc
// @Summary This controller will handles to fetch profile of students.
// @description This function retrieves the student's profile based on the provided Authorization token and
// @description fetches the student's demo session details. It responds with a JSON object containing the student's information
// @description and demo session details.
// @Tags Student
// @Accept application/x-www-form-urlencoded
// @Param userID query  number true "User ID"
// @Produce json
// @Success 200
// @Router /profile [get]
func GetUserProfile(c *gin.Context) {
	// Fetch the authorization token from the request header.
	userIDStr := c.Query("userID")
	if len(userIDStr) == 0 {
		log.Println("GetUserProfile: Failed to fetch user's id.")
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Failed",
			"message": "Failed to fetch user id from query params.",
		})
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Println("GetUserProfile: Failed to parse user's id into integer.")
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Failed",
			"message": "Failed to parse user id into integer.",
		})
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)

	var userProfile models.User
	var fetchProfileErr error

	go func() {
		defer wg.Done()
		// Fetch the student's profile details using user ID.
		userProfile, fetchProfileErr = models.GetUserProfile(userID)
		if fetchProfileErr != nil {
			log.Println("[ERROR] GetUserProfile: Failed to fetch user's details by using user ID with error: ", fetchProfileErr)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "Failed",
				"message": "Failed to fetch user details from database.",
			})
			return
		}
	}()
	wg.Wait()
	// Send the JSON response with the student profile and demo session details.
	c.JSON(http.StatusOK, userProfile)

}
