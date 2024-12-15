package models

import (
	"log"
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
    VALUES ($1, $2, $3, $4, $5, $6,$7,$8,$9)
			ON CONFLICT (phone_number)
			DO UPDATE SET
			otp = $2,
			otp_valid_until = $3
			RETURNING id,phone_verified,email_verified,referral_code,query_param,
			 CASE WHEN xmax = 0 THEN 'insert' ELSE 'update' END AS action`
	log.Println("===", query)
	var (
		user_id       int
		phoneVerified bool
		action        string
	)
	userLocation := user.Location.City + ", " + user.Location.State + ", " + user.Location.Country
	otpValidUntil := time.Now().Add(time.Minute * 5)
	// Set phone_otp_expire time to 24 hrs from date of student generated.

	err = db.QueryRow(query, user.Phone, user.OTP, otpValidUntil, userLocation, user.UserIP, user.DialingCode).Scan(&user_id, &phoneVerified, &action)

	if err != nil {
		log.Println("SaveOTP: failed while execute the query for saving otp in database with error :", err)
		return "", err
	}

	user.ID = int64(user_id)
	user.IsPhoneVerified = phoneVerified

	return action, nil
}
