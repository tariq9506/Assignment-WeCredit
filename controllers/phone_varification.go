package controllers

import (
	"log"
	"os"
	"we-credit/models"
	"we-credit/service"
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
