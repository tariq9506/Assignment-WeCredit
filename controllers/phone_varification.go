package controllers

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"we-credit/models"
	"we-credit/service"
	"we-credit/utility"

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
// @description This api is taking code and userId as postform
// @Tags PhoneVerification
// @Accept application/x-www-form-urlencoded
// @Param code  formData  string true "Code"
// @Param user-id  formData  string true "user-id"
// @Param User-Agent header string false "User Agent"
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
	log.Println("CODE ENTER BY user: ", code)

	userID, err := strconv.Atoi(c.PostForm("user-id"))
	if err != nil {
		log.Println("Verification Code Error: Failed to convert user ID into a valid number.", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Failed",
			"message": "Please enter a valid user id",
		})
		return
	}

	// function use to fetch user details by user id
	user, err := models.GetUserByID(userID)
	if err != nil {
		log.Println("VerifyCode: failed to fetch user data :", err)
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

// ResendVerificationCode godoc
// @Summary This controller will resend the given code same as OTP.
// @description This controller will resend the OTP.
// @description This api is taking phone number as postform and user ip from header.
// @Tags PhoneVerification
// @Accept application/x-www-form-urlencoded
// @Param phone-number  formData  string true "Phone Number"
// @Produce json
// @Success 200
// @Router /otp/send [POST]
func ResendVerificationCode(c *gin.Context) {

	userIP := utility.GetClientIP(c)
	phoneNumber := c.PostForm("phone-number")
	if len(phoneNumber) == 0 {
		log.Println("ResendVerificationCode: failed, phone nnumber can not be empty.")
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Failed",
			"message": "Please enter a valid phone no.",
		})
		return
	}
	match, err := regexp.MatchString("^[0-9]{10}$", phoneNumber)
	if !match || err != nil {
		log.Println("ResendVerificationCode: Failed, invalid phone number with error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Failed",
			"message": "Please enter a valid phone no.",
		})
		return
	}
	location := service.GetLocationFromIP(userIP)
	details, err := models.GetDetailsOfSupportedCountryByCode(location.CountryCode)
	if err != nil {
		log.Println("ResendVerificationCode: GetDetailsOfSupportedCountryByCode failed to get location iformation with error: ", err)
	}
	otp := utility.GenerateOTP()

	user := models.User{
		Phone:       phoneNumber,
		DialingCode: details.CountryPhoneCode,
		UserIP:      userIP,
		Location:    location,
		OTP:         otp,
	}
	_, err = models.SaveOTP(user)
	if err != nil {
		log.Println("ResendVerificationCode Failed: Unable to send verification code. Please try again later", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Failed",
			"message": "Failed to save otp",
		})
		return
	}
	// func to send the verification code to the user's phone number
	err = SendPhoneNumberVerificationCode(user)
	if err != nil {
		log.Println("ResendVerificationCode Failed: Unable to send verification code. Please try again later", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Failed",
			"message": "Failed to send otp",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"mesage": "One time message has been sent to you phone number",
	})
}
