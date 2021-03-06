package controllers

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	cfg "github.com/adopabianko/p2p-auth/config"
)

type Register struct {
	GroupID              int8   `json:"group_id"`
	Name                 string `json:"name"`
	CompanyName          string `json:"company_name"`
	Gender               int8   `json:"gender"`
	BirthDate            string `json:"birth_date"`
	JobID                int8   `json:"job_id"`
	Address              string `json:"address"`
	ProvinceID           int    `json:"province_id"`
	CityID               int    `json:"city_id"`
	PhoneNumber          string `json:"phone_number"`
	IdentityType         int8   `json:"identity_type"`
	IdentityFile         string `json:"identity_file"`
	NpwpFile             string `json:"npwp_file"`
	SiupFile             string `json:"siup_file"`
	Email                string `json:"email"`
	Password             []byte `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

type Login struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

type VerificationAccount struct {
	VerificationCode string `json:"verification_code"`
}

// Default Router
func IndexPage(w http.ResponseWriter, r *http.Request) {
	res := map[string]string{"message": "Membership Api", "version": "V.0.0.1"}
	json, err := json.Marshal(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	var reg Register

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &reg)

	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		res := map[string]interface{}{"status": "Error", "message": "Error validation", "data": err.Error()}
		json, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)

	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		res := map[string]interface{}{"status": "Error", "message": "Error hash", "data": err.Error()}
		json, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	}

	if reg.GroupID == 2 { // Borrower
		dataRegisterBorrower := map[string]interface{}{
			"groupID": reg.GroupID,
			"email": reg.Email,
			"password": hashedPassword,
			"name": reg.Name,
			"companyName": reg.CompanyName,
			"address": reg.Address,
			"provinceID": reg.ProvinceID,
			"cityID": reg.CityID,
			"phoneNumber": reg.PhoneNumber,
			"identityType": reg.IdentityType,
			"identityFile": reg.IdentityFile,
			"npwpFile": reg.NpwpFile,
			"siupFile": reg.SiupFile,
		} 
		err = registerBorrower(dataRegisterBorrower)

		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			res := map[string]interface{}{"status": "Error", "message": "Database error", "data": err.Error()}
			json, _ := json.Marshal(res)

			w.Header().Set("Content-Type", "application/json")
			w.Write(json)
			return
		}
	} else if reg.GroupID == 3 { // Investor
		dataRegisterInvestor := map[string]interface{}{
			"groupID": reg.GroupID,
			"email": reg.Email,
			"password": hashedPassword,
			"name": reg.Name,
			"gender": reg.Gender,
			"birthDate": reg.BirthDate,
			"jobID": reg.JobID,
			"address": reg.Address,
			"provinceID": reg.ProvinceID,
			"cityID": reg.CityID,
			"phoneNumber": reg.PhoneNumber,
			"identityType": reg.IdentityType,
			"identityFile": reg.IdentityFile,
		}
		err = registerInvestor(dataRegisterInvestor)

		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			res := map[string]interface{}{"status": "Error", "message": "Database error", "data": err.Error()}
			json, _ := json.Marshal(res)

			w.Header().Set("Content-Type", "application/json")
			w.Write(json)
			return
		}
	}

	res := map[string]interface{}{"status": "Ok", "message": "success", "data": nil}
	json, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func registerBorrower(dataRegisterBorrower map[string]interface{}) error {
	var verificationCode string = generateVerificationCode()

	db := cfg.DBConnection()

	qUserAccount := `INSERT INTO user_accounts
		(group_id, email, password, verification_code) VALUES ($1, $2, $3, $4)
		RETURNING id`

	var userAccountID int

	err := db.QueryRow(qUserAccount, dataRegisterBorrower["groupID"], dataRegisterBorrower["email"], dataRegisterBorrower["password"], verificationCode).Scan(&userAccountID)

	qClient := `INSERT INTO clients
		(user_account_id, name, company_name, address, province_id, city_id, phone_number, identity_type, identity_file, npwp_file, siup_file)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = db.Query(qClient, userAccountID, dataRegisterBorrower["name"], dataRegisterBorrower["companyName"], dataRegisterBorrower["address"], dataRegisterBorrower["provinceID"], dataRegisterBorrower["cityID"], dataRegisterBorrower["phoneNumber"], dataRegisterBorrower["identityType"], dataRegisterBorrower["identityFile"], dataRegisterBorrower["npwpFile"], dataRegisterBorrower["siupFile"])
	defer db.Close()

	return err
}

func registerInvestor(dataRegisterInvestor map[string]interface{}) error {
	var verificationCode string = generateVerificationCode()

	db := cfg.DBConnection()

	qUserAccount := `INSERT INTO user_accounts
		(group_id, email, password, verification_code) VALUES ($1, $2, $3, $4)
		RETURNING id`

	var userAccountID int

	err := db.QueryRow(qUserAccount, dataRegisterInvestor["groupID"], dataRegisterInvestor["email"], dataRegisterInvestor["password"], verificationCode).Scan(&userAccountID)

	qClient := `INSERT INTO clients
		(user_account_id, name, gender, birth_date, job_id, address, province_id, city_id, phone_number, identity_type, identity_file)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = db.Query(qClient, userAccountID, dataRegisterInvestor["name"], dataRegisterInvestor["gender"], dataRegisterInvestor["birthDate"], dataRegisterInvestor["jobID"], dataRegisterInvestor["address"], dataRegisterInvestor["provinceID"], dataRegisterInvestor["cityID"], dataRegisterInvestor["phoneNumber"], dataRegisterInvestor["identityType"], dataRegisterInvestor["identityFile"])
	defer db.Close()

	return err
}

func VerificationAccountPage(w http.ResponseWriter, r *http.Request) {
	var va VerificationAccount

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &va)

	db := cfg.DBConnection()
	qAccount := `SELECT verification_code FROM user_accounts WHERE verification_code = $1 AND status = 0`
	err = db.QueryRow(qAccount, va.VerificationCode).Scan(&va.VerificationCode)

	if err != nil {
		defer db.Close()

		res := map[string]interface{}{"status": "Error", "message": err.Error(), "data": nil}
		json, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	}

	qUpdateAccount := `UPDATE user_accounts SET status = 1 WHERE verification_code = $1`
	_, err = db.Query(qUpdateAccount, va.VerificationCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer db.Close()

	res := map[string]interface{}{"status": "Ok", "message": "success", "data": nil}
	json, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	var login Login
	var email string
	var password []byte

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &login)

	db := cfg.DBConnection()
	qAccount := `SELECT email, password FROM user_accounts WHERE email = $1 AND status = 1`
	err = db.QueryRow(qAccount, login.Email).Scan(&email, &password)

	if err != nil {
		defer db.Close()

		res := map[string]interface{}{"status": "Error", "message": "User not found", "data": nil}
		json, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(login.Password))

	if err != nil {
		defer db.Close()

		res := map[string]interface{}{"status": "Error", "message": "User not found", "data": nil}
		json, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	}

	defer db.Close()

	res := map[string]interface{}{"status": "Ok", "message": "success", "data": nil}
	json, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func CheckUserAccountPage(w http.ResponseWriter, r *http.Request) {
	var countUser int8

	GroupID := r.URL.Query().Get("group_id")
	Email := r.URL.Query().Get("email")

	db := cfg.DBConnection()
	qAccount := `SELECT count(*) FROM user_accounts WHERE group_id = $1 AND email = $2 AND status = 1`
	err := db.QueryRow(qAccount, GroupID, Email).Scan(&countUser)

	// if database error or no result data
	if err != nil {
		defer db.Close()

		res := map[string]interface{}{"status": "Error", "message": err.Error(), "data": nil}
		json, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	} else if countUser != 0 {
		defer db.Close()

		res := map[string]interface{}{"status": "Error", "message": "User already exists", "data": nil}
		json, _ := json.Marshal(res)

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		return
	}

	defer db.Close()

	res := map[string]interface{}{"status": "Ok", "message": "success", "data": nil}
	json, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func generateVerificationCode() string {
	var len int = 6

	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25))
	}
	return string(bytes)
}
