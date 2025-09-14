package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"restapi/internal/models"
	"restapi/internal/repository/sqlconnect"
	"restapi/pkg/utils"
	"strconv"
)

func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	var teachers []models.Teacher
	teachers, err := sqlconnect.GetTeachersDBHandler(teachers, r)
	if err != nil {
		return
	}

	response := struct{
		Status string `json:"status"`
		Count int `json:"count"`
		Data []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count: len(teachers),
		Data: teachers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	
}

func GetTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	// Handle Path parameter
	id, err := strconv.Atoi(idStr) 
	if err != nil {
		fmt.Println(err)
		return
	}

	teacher, err := sqlconnect.GetTeacherByID(id)
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {
	var newTeachers []models.Teacher
	var rawTeachers []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawTeachers)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fields := GetFieldNames(models.Teacher{})

	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, teacher := range rawTeachers {
		for key := range teacher {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable field found in request, Only use allowed fields.", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newTeachers)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, teacher := range newTeachers {
		err := CheckBlankFields(teacher)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	addedTeachers, err := sqlconnect.AddTeachersDBHandler(newTeachers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string `json:"status"`
		Count int `json:"count"`
		Data []models.Teacher `json:"data"`
	} {
		Status: "success",
		Count: len(addedTeachers),
		Data: addedTeachers,
	}

	json.NewEncoder(w).Encode(response)
}

func UpdateTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher Id", http.StatusBadRequest)
		return
	}

	var updatedTeacher models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	updatedTeacherFromDB, err := sqlconnect.UpdateTeacher(id, updatedTeacher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacherFromDB)

}

func PatchTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher Id", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	updatedTeacherFromDB, err := sqlconnect.PatchTeacher(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacherFromDB)
}

func PatchTeachersHandler(w http.ResponseWriter, r *http.Request) {
	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = sqlconnect.PatchTeachers(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func DeleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher Id", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DeleteOneTeacher(id)
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
		Status: "Teacher Successfully deleted",
		ID: id,
	}
	json.NewEncoder(w).Encode(response)
}

func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {
	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	deletedIds, err := sqlconnect.DeleteTeachers(ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		DeletedIds []int `json:"deleted_ids"`
	} {
		Status: "Teachers Successfully deleted",
		DeletedIds: deletedIds,
	}
	json.NewEncoder(w).Encode(response)
}

func GetStudentsByTeacherId (w http.ResponseWriter, r *http.Request) {
	teacherId := r.PathValue("id")

	var students []models.Student

	students, err := sqlconnect.GetStudentsByTeacherIdFomDB(teacherId, students)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	response := struct {
		Status string `json:"status"`
		Count int `json:"count"`
		Data []models.Student `json:"data"`
	} {
		Status: "success",
		Count: len(students),
		Data: students,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetStudentCountByTeacherId(w http.ResponseWriter, r *http.Request) {
	// admin, manager, exec
	_, err := utils.AuthorizeUser(r.Context().Value(utils.ContextKey("role")).(string), "admin", "manager", "exec")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	teacherId := r.PathValue("id")

	var studentCount int

	studentCount, err = sqlconnect.GetStudentCountByTeacherIdFromDB(teacherId)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	response := struct {
		Status string `json:"status"`
		Count int `json:"count"`
	} {
		Status: "success",
		Count: studentCount,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}