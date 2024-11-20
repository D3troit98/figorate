package helpers

import (
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a hashed password with a plain one
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func ValidatePassword(password,email string) error{
	//  Check if password contain spaces
    if strings.Contains(password, " "){
        return errors.New("password cannot contain spaces")
    }

    // Check password length
    if len(password) < 8{
        return errors.New("password ust be at least 8 character")
    }

    // Check for at least one symbol or number
    symbolOrNumberRegex := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>0-9]`)
    if !symbolOrNumberRegex.MatchString(password){
        return errors.New("password must contain at least one symbol or number")
    }

    // Check if password contains email or parts of email
    lowercasePassword := strings.ToLower(password)
    lowercaseEmail := strings.ToLower(email)
    emailParts := strings.Split(lowercaseEmail, "@")

    // Check against full email
    if strings.Contains(lowercasePassword, lowercaseEmail){
        return errors.New("password cannot include your email address")
    }


    // Check against email username(before @)
    if len(emailParts) > 0 && strings.Contains(lowercasePassword, emailParts[0]){
        return errors.New("password cannot include your email username")
    }
    return nil
}
