package controllers

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"we-credit/models"
	"we-credit/service"

	"github.com/gin-gonic/gin"
)

// SendPhoneNumberVerificationCode generates an OTP, saves it to the database, and sends it via SMS to the specified phone number.
// Parameters:
// - phoneNo: The phone number to which the OTP will be sent.
// - dialingCode: The international dialing code for the phone number.
// - userIP: Information of user's IP.
// - location: Location contain user's location information.
// Returns:
// - error: Any error encountered during the process.
func SendPhoneNumberVerificationCode(user models.User) error {

	var message string
	if user.IsPhoneVerified {
		message = user.OTP + " is the verification code to log in to your Tutree account. Please DO NOT SHARE this code with anyone.\n@" + os.Getenv("DOMAIN_NAME") + " #" + user.OTP

	} else {
		// content for the otp message
		message = user.OTP + " is the verification code to sign up to your Tutree account. Please DO NOT SHARE this code with anyone.\n@" + os.Getenv("DOMAIN_NAME") + " #" + user.OTP
	}

	log.Println("OTPmessage", message)
	// This is the twilio service to send the otp to the given phone number.
	// function use to send message given phone having message and otp
	err := service.SendMessage(user.Phone, message, user.DialingCode)
	if err != nil {
		log.Println("SendPhoneNumberVerificationCode: sending otp failed: ", err)
		return err
	}

	return nil
}

// VerifyCode godoc
// @Summary This controller will verify the given code same as OTP. It will also check if OTP is expired
// @description This controller will verify the given code same as OTP. It will also check if OTP is expired.
// @description This api is taking code and studentId as postform
// @Tags PhoneVerification
// @Accept application/x-www-form-urlencoded
// @Param code  formData  string true "Code"
// @Param user-id  formData  string true "user-id"
// @Param User-Agent header string true "User Agent"
// @Produce json
// @Success 200
// @Router /otp/verify [POST]
func VerifyCode(c *gin.Context) {

	code := c.PostForm("code")
	if len(code) == 0 {
		log.Println("VerifyCode: Verification Code Required, Please enter the OTP (One-Time Password) to proceed.")
		c.JSON(http.StatusOK, gin.H{
			"status":  "Failed",
			"message": "Please enter a valid code.",
		})
		return

	}
	log.Println("CODE ENTER BY STUDENT: ", code)

	userID, err := strconv.Atoi(c.PostForm("user-id"))
	if err != nil {
		log.Println("Verification Code Error: Failed to convert student ID into a valid number.", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Failed",
			"message": "Please enter a valid user id",
		})
		return
	}

	// function use to fetch student details by student id
	user, err := models.GetUserByID(userID)
	if err != nil {
		log.Println("VerifyCode: failed to fetch student data :", err)
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Failed",
			"message": "Please enter a valid user details.",
		})
		return
	}
	OTP, err := models.GetValidVerificationCode(int64(userID))
	if err != nil {
		log.Println("VerifyCode: Error occurred while fetching OTP or checking if it is expired for user ID:", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Failed",
			"message": "Please enter a valid otp, OTP expired.",
		})
		return
	}
	// for local, testing environment this line of code would accept "1234" as valid OTP.
	if code == OTP {
		err := models.SetPhoneVerified(int64(userID))
		if err != nil {
			log.Println("VerifyCode: failed to verify Phone number:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "Failed",
				"message": "Please enter a valid user ID.",
			})
			return
		}

		token := CreateUserAuth(c, user)

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Sucessfully verified phone number.",
			"token":   token,
			"user_id": user.ID,
		})
		return
	} else {
		log.Println("Verification Failed: The OTP entered is incorrect. Please retry with the correct OTP.")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Failed",
			"message": "Please enter a valid otp",
		})
		return

	}
}
