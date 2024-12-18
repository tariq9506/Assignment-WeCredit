package models

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"
	"we-credit/config"
	"we-credit/service"
)

type User struct {
	ID              int64            `json:"id,omitempty"`
	Phone           string           `json:"phone,omitempty"`
	DialingCode     string           `json:"dialing_code,omitempty"`
	OTP             string           `json:"otp,omitempty"`
	OTPValidUntil   time.Time        `json:"otp_valid_until,omitempty"`
	UserIP          string           `json:"user_ip,omitempty"`
	Location        service.Location `json:"location,omitempty"`
	IsPhoneVerified bool             `json:"IsPhoneVerified,omitempty"`
	CreatedAt       time.Time        `json:"created_at,omitempty"`
}

func SaveNewUser(user *User) (string, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("SaveOTP: Failed while connecting with the database :", err)
		return "", err
	}
	defer db.Close()

	query := `
			INSERT INTO public.user (phone_number, otp,otp_valid_until,location, ip,dialing_code)
    VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (phone_number)
			DO UPDATE SET
			otp = $2,
			otp_valid_until = $3
			RETURNING id,phone_verified,
			 CASE WHEN xmax = 0 THEN 'insert' ELSE 'update' END AS action`
	log.Println("===", query)
	var (
		user_id       int
		phoneVerified bool
		action        string
	)
	userLocation := user.Location.City + ", " + user.Location.State + ", " + user.Location.Country
	otpValidUntil := time.Now().Add(time.Minute * 5)
	// Set phone_otp_expire time to 24 hrs from date of user generated.

	err = db.QueryRow(query, user.Phone, user.OTP, otpValidUntil, userLocation, user.UserIP, user.DialingCode).Scan(&user_id, &phoneVerified, &action)

	if err != nil {
		log.Println("SaveOTP: failed while execute the query for saving otp in database with error :", err)
		return "", err
	}

	user.ID = int64(user_id)
	user.IsPhoneVerified = phoneVerified

	return action, nil
}

// GetUserByID retrieves a user record from the database by user ID.
// Parameters:
// - userID: The ID of the user to retrieve.
// Returns:
// - user: The user record retrieved from the database.
// - error: Any error encountered during the process.
func GetUserByID(userID int) (User, error) {

	query := `
		SELECT
			id, 
			phone_number,
			phone_verified,
			otp,
			location
		FROM 
			public.user 
		WHERE 
			id=` + strconv.Itoa(userID)

	return getUser(query)
}
func getUser(query string) (User, error) {

	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetUserByID: Failed while connecting with the database :", err)
		return User{}, err
	}
	defer db.Close()
	var (
		ID              sql.NullInt64
		phone           sql.NullString
		isPhoneVerified sql.NullBool
		otp             sql.NullString
		location        sql.NullString
		user            User
	)

	err = db.QueryRow(query).Scan(
		&ID,
		&phone,
		&isPhoneVerified,
		&otp,
		&location,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("GetUserByID : user doesnt exist error: ", err)
		}
	}

	var locStruct service.Location
	loc := strings.Split(location.String, ", ")
	if len(loc) > 2 {
		locStruct.City = loc[0]
		locStruct.State = loc[1]
		locStruct.CountryCode = loc[2]

	}

	user = User{
		ID:              ID.Int64,
		Phone:           phone.String,
		IsPhoneVerified: isPhoneVerified.Bool,
		OTP:             otp.String,
		Location:        locStruct,
	}

	return user, err
}

// GetuserProfile retrieves the profile details of a user from the database based on the provided user ID.
// It returns a user struct containing the user's information or an error if the operation fails.
//
// Parameters:
//   - userID (int): The unique identifier for the user whose profile is to be retrieved.
//
// Returns:
//   - user: A struct containing the user's profile details such as name, grade, parent email, picture URL, and school ID.
//   - error: An error object if there is any failure during database connection or query execution.
func GetUserProfile(userID int) (User, error) {
	db, err := config.GetDB2()
	var userDetails User
	if err != nil {
		log.Println("GetUserProfile: Failed when try to connect with database with error: ", err)
		return userDetails, err
	}
	defer db.Close()
	var (
		isPhoneVerified sql.NullBool
		phoneNumber     sql.NullString
		dialingCode     sql.NullString
		location        sql.NullString
		createdAt       sql.NullTime
	)
	query := `
		SELECT                                     
    			u.phone_verified,
    			u.phone_number,
    			u.dialing_code,
    			u.location,
				u.created_at
		FROM 
		    public.user AS u 
		WHERE 
   			 u.id = $1
`

	err = db.QueryRow(query, userID).Scan(
		&isPhoneVerified,
		&phoneNumber,
		&dialingCode,
		&location,
		&createdAt,
	)
	if err == sql.ErrNoRows {
		return userDetails, nil
	}
	if err != nil {
		log.Println("GetUserProfile: Failed while execute the query with error: ", err)
		return userDetails, err
	}
	var locStruct service.Location
	loc := strings.Split(location.String, ", ")
	if len(loc) > 2 {
		locStruct.City = loc[0]
		locStruct.State = loc[1]
		locStruct.CountryCode = loc[2]

	}
	// Create a map to store the session details.
	userDetails = User{
		IsPhoneVerified: isPhoneVerified.Bool,
		Phone:           phoneNumber.String,
		DialingCode:     dialingCode.String,
		Location:        locStruct,
		CreatedAt:       createdAt.Time,
	}
	return userDetails, nil
}
