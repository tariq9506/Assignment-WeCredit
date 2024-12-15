package models

import (
	"database/sql"
	"errors"
	"log"
	"time"
	"we-credit/config"
)

// GetValidVerificationCode
// input : userID
// Output: OTP code, error
// Desc  : This function will return the OTP code, and it will also check current time is less than the expire time of OTP.
func GetValidVerificationCode(userID int64) (string, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetValidVerificationCode: Failed while connecting with the database :", err)
		return "", err
	}
	defer db.Close()

	var OTP sql.NullString
	var OTP_expire sql.NullTime

	err = db.QueryRow("SELECT otp, otp_valid_until FROM public.user WHERE id=$1", userID).Scan(&OTP, &OTP_expire)
	if err != nil {
		log.Println("GetValidVerificationCode: Failed while querying and scanning the row:", err)
		return "", err
	}

	// Checking the OTP expire
	expireDate := (OTP_expire).Time.Format("2006-01-02 15:04:05")
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	l_expireDate, _ := time.Parse("2006-01-02 15:04:05", expireDate)
	l_currentTime, _ := time.Parse("2006-01-02 15:04:05", currentTime)

	result := l_currentTime.Before(l_expireDate)

	if !result {
		otp_error := errors.New("OTP is expired")
		return "", otp_error
	}

	return (OTP).String, nil
}

// SetPhoneVerified
// input : studentID, verification flag
// Output: error
// Desc  : This function will set the phone verification flag as true/false.
func SetPhoneVerified(userID int64) error {

	db, err := config.GetDB2()
	if err != nil {
		log.Println("SetPhoneVerified: Failed while connecting with the database :", err)
		return err
	}
	defer db.Close()

	query := `
				UPDATE
					public.user
				SET
					phone_verified = true
				WHERE id=$1`

	_, err = db.Exec(query, userID)
	if err != nil {
		log.Println("SetPhoneVerified: failed while execute the query with error ", err)
		return err
	}

	return nil
}

// SaveOTP
// input : phone number, OTP,user ip , location
// Output: student struct or error
// Desc  : This function will save the OTP and expire time of OTP in database.
func SaveOTP(user User) (User, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("SaveOTP: Failed while connecting with the database :", err)
		return User{}, err
	}
	defer db.Close()

	query := `
			INSERT INTO public.user (phone_number, otp,otp_valid_until,`
	query += `location, ip)
    VALUES ($1, $2, $3,$4,$5 `

	query += `)
			ON CONFLICT (phone_number)
			DO UPDATE SET
			otp = $2,
			otp_valid_until = $3
			RETURNING id,phone_verified`
	log.Println("===", query)
	var (
		user_id       int
		phoneVerified bool
	)
	userLocation := user.Location.City + ", " + user.Location.State + ", " + user.Location.Country
	otpValidUntil := time.Now().Add(time.Minute * 5)
	// Set phone_otp_expire time to 24 hrs from date of student generated.

	err = db.QueryRow(query, user.Phone, user.OTP, otpValidUntil, userLocation, user.UserIP).Scan(&user_id, &phoneVerified)
	if err != nil {
		log.Println("SaveOTP: failed while execute the query for saving otp in database with error :", err)
		return User{}, err
	}
	return User{
		ID:              int64(user_id),
		IsPhoneVerified: phoneVerified,
	}, nil
}
