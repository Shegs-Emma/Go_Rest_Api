package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"restapi/internal/models"
	"restapi/internal/repository/sqlconnect"
	"restapi/pkg/utils"
	"strconv"
	"time"
)

func GetExecsHandler(w http.ResponseWriter, r *http.Request) {
	var execs []models.Exec
	execs, err := sqlconnect.GetExecsDBHandler(execs, r)
	if err != nil {
		return
	}

	response := struct{
		Status string `json:"status"`
		Count int `json:"count"`
		Data []models.Exec `json:"data"`
	}{
		Status: "success",
		Count: len(execs),
		Data: execs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	
}

func GetExecHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	// Handle Path parameter
	id, err := strconv.Atoi(idStr) 
	if err != nil {
		fmt.Println(err)
		return
	}

	exec, err := sqlconnect.GetExecByID(id)
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exec)
}

func AddExecsHandler(w http.ResponseWriter, r *http.Request) {
	var newExecs []models.Exec
	var rawExecs []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawExecs)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusForbidden)
		return
	}

	fields := GetFieldNames(models.Exec{})

	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, exec := range rawExecs {
		for key := range exec {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable field found in request, Only use allowed fields.", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newExecs)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, exec := range newExecs {
		err := CheckBlankFields(exec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	addedExecs, err := sqlconnect.AddExecsDBHandler(newExecs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string `json:"status"`
		Count int `json:"count"`
		Data []models.Exec `json:"data"`
	} {
		Status: "success",
		Count: len(addedExecs),
		Data: addedExecs,
	}

	json.NewEncoder(w).Encode(response)
}

func PatchExecHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid exec Id", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	updatedExecFromDB, err := sqlconnect.PatchExec(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedExecFromDB)
}

func PatchExecsHandler(w http.ResponseWriter, r *http.Request) {
	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = sqlconnect.PatchExecs(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func DeleteExecHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Exec Id", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DeleteOneExec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// w.WriteHeader(http.StatusNoContent)
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID int `json:"id"`
	} {
		Status: "Exec Successfully deleted",
		ID: id,
	}
	json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Exec

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	user, err := sqlconnect.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "invalid username or password", http.StatusBadRequest)
		return
	}

	if user.InactiveStatus {
		http.Error(w, "account is inactive", http.StatusForbidden)
		return
	}

	err = utils.VerifyPassword(req.Password, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	tokenString, err := utils.SignToken(user.ID, req.Username, user.Role)
	if err != nil {
		utils.ErrorHandler(errors.New("token could not be generated"), "token could not be generated")
		http.Error(w, "token could not be generated", http.StatusForbidden)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "Bearer",
		Value: tokenString,
		Path: "/",
		HttpOnly: true,
		Secure: true,
		Expires: time.Now().Add(24*time.Hour),
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name: "test",
		Value: "testing",
		Path: "/",
		HttpOnly: true,
		Secure: true,
		Expires: time.Now().Add(24*time.Hour),
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Token string `json:"token"`
	} {
		Token: tokenString,
	}
	json.NewEncoder(w).Encode(response)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: "Bearer",
		Value: "",
		Path: "/",
		HttpOnly: true,
		Secure: true,
		Expires: time.Unix(0, 0),
		SameSite: http.SameSiteStrictMode,
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Logged out successfully"}`))
}

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid exec id", http.StatusBadRequest)
		return
	}

	var req models.UpdatePasswordRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "Please enter password", http.StatusBadRequest)
		return
	}

	_, err = sqlconnect.UpdatePasswordInDB(userId, req.CurrentPassword, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Message string `json:"message"`
	} {
		Message: "Password updated successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return 
	}
	r.Body.Close()

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	err = sqlconnect.ForgotPasswordDbHandler(req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Password reset link sent to %s", req.Email)
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("resetcode")

	type request struct {
		NewPassword string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	var req request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid values in request", http.StatusBadRequest)
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		http.Error(w, "Passwords should match", http.StatusBadRequest)
		return
	}

	err = sqlconnect.ResetPasswordDbHandler(token, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, "Password reset successfully")
}