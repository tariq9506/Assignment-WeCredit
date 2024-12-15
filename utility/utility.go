package utility

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetHostURL get host url
func GetHostURL() string {
	return os.Getenv("HOST_URL")
}
func GetClientIP(c *gin.Context) string {
	clientIP := c.ClientIP()
	if os.Getenv("ENV") == "local" {
		clientIP = os.Getenv("LOCAL_IP")
	}
	return clientIP
}

// input: no parameter
// output: bool
// func GetTwilioAccountID will check whether voip numbers on the server is allowed or not.
func AllowVoipNumbers() bool {
	isAllowed, err := strconv.ParseBool(os.Getenv("ALLOW_VOIP_NUMBERS"))
	if err != nil {
		log.Println("[ENV-MISSING] AllowVoipNumbers: failed while fetching .env variable for voip number")
		return true
	}
	return isAllowed
}

// input: no parameter
// output: string
// func GetTwilioAccountID will return twilio account sid which will be used for calling twilio api service
func GetTwilioAccountID() string {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	return accountSid
}

// input: no parameter
// output: string
// func GetTwilioAccountID will return twilio account authorization token which will be used for calling twilio api service
func GetTwilioAuthorizationToken() string {
	authToken := os.Getenv("TWILIO_ACCOUNT_AUTH_TOKEN")
	return authToken
}

// generateOTP
// input :
// Output: OTP
// Desc  : This controller will generate OTP.
// OTP Generation
func GenerateOTP() string {
	charSet := "1234567890"
	otp := randomStringGenerator(charSet, 4)
	return otp
}
func randomStringGenerator(charSet string, codeLength int32) string {
	code := ""
	charSetLength := int32(len(charSet))

	// Seed the random number generator to ensure different results each time
	randomNumber := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := int32(0); i < codeLength; i++ {
		// Generate a random index within the bounds of charSetLength
		index := randomNumber.Intn(int(charSetLength))

		// Append the character at the generated index to the code
		code += string(charSet[index])
	}

	// Return the generated random string
	return code
}
