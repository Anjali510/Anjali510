package main

import (
	//"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gorilla/mux"
	//"gopkg.in/go-playground/validator.v9"
)

type Question_Details struct {
	QuestionInformation Question_Information `json:"StudentInformation"`
	Suggestion          []Suggestions        `json:"ParentInformation"`
}

type Question_Information struct {
	ID          *int     `json:"ID"`
	Name        *string  `json:"Name"`
	Description *string  `json:"Description"`
	Price       *float64 `json:"Price"`
}

type Suggestions struct {
	Suggest *string `json:"Suggest"`
}

type StandardResponse struct {
	Data  string `json:"Data"`
	Value int    `json:"Value"`
}

var myAddress string = ""
var connString string

func list(w http.ResponseWriter, r *http.Request) {
	//Set header type and initialize connection string

	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("mssql", connString)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("List api called")

	defer r.Body.Close()

	var input *Question_Details = &Question_Details{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err.Error())
	}

	//Decode input json and insert it into s
	err = json.Unmarshal(body, &input)

	if err != nil {
		panic(err.Error())
	}

	//The filters will be stored in vals now
	filters := ""
	filterApplied := false

	//Query to insert student information data
	query := "select ID, Name, Description, Price where"

	vals := []interface{}{}
	if input.QuestionInformation.Name != nil {
		filters += "Name = ? and "
		vals = append(vals, input.QuestionInformation.Name)
		filterApplied = true
	}
	if input.QuestionInformation.ID != nil {
		filters += "ID = ? and "
		vals = append(vals, input.QuestionInformation.ID)
		filterApplied = true
	}

	if input.QuestionInformation.Description != nil {
		filters += "Description = ? and "
		vals = append(vals, input.QuestionInformation.Description)
		filterApplied = true
	}

	if input.QuestionInformation.Price != nil {
		filters += "Price = ? and "
		vals = append(vals, input.QuestionInformation.Price)
		filterApplied = true
	}

	if filterApplied {
		filters = filters[0 : len(filters)-5]
		filters += ";"
		query += filters
	} else {
		query = query[0 : len(query)-7]
		query += ";"
	}

	results, err := db.Query(query, vals...)
	if err != nil {
		panic(err.Error())
	}

	var students []Question_Details
	for results.Next() {
		var student Question_Details
		err := results.Scan(&student.QuestionInformation.ID, &student.QuestionInformation.Name, &student.QuestionInformation.Description, &student.QuestionInformation.Price)
		if err != nil {
			panic(err.Error())
		}

		// Parent Filter
		saquery := "select Suggest from sis_suggestions where ID=?;"

		saresults, err := db.Query(saquery, *&student.QuestionInformation.ID)

		if err != nil {
			panic(err.Error())
		}

		for saresults.Next() {
			var parent Suggestions
			err := saresults.Scan(&parent.Suggest)
			if err != nil {
				panic(err.Error())
			}
			student.Suggestion = append(student.Suggestion, parent)
		}

		students = append(students, student)

	}
	json.NewEncoder(w).Encode(students)
}

func handler() {
	router := mux.NewRouter()
	router.HandleFunc("/list", list).Methods("POST")
}

func main() {
	//Open ServerLocations File containing location of jsons
	jsonFile, err := os.Open("./ServerLocations.json")
	if err != nil {
		panic(err.Error())
	}

	//Read bytes from the opened file
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err.Error())
	}

	var results []map[string]interface{}
	err = json.Unmarshal(byteValue, &results)
	if err != nil {
		panic(err.Error())
	}

	for _, result := range results {
		if result["Name"] == "QuestionInfo" {
			str, ok := result["message"].(string)
			if ok {
				myAddress = str
			} else {
				panic(err.Error())
			}

			str, ok = result["DBConnectionString"].(string)
			if ok {
				connString = str
			} else {
				panic(err.Error())
			}
		}
	}

	jsonFile.Close()

	handler()
}
