package dal

// Import required packages
import (
	"encoding/json"                    // For JSON handling
	"fmt"                              // For formatted I/O
	_ "github.com/go-sql-driver/mysql" // Import mysql driver
	"log"                              // For logging
	"time"                             // For simulating machine learning model processing time
)

// Prediction struct models the data structure of a prediction in the database
//
// This code defines a struct named "Prediction" with fields for PredictionID, EngineID, InputData, PredictionInfo, and PredictionTime.
type Prediction struct {
	PredictionID   string
	EngineID       string
	InputData      string
	PredictionInfo string
	PredictionTime string
}

// Function to check if the engine_id exists in scraper_engine table
//
// This function checks if a given engine ID exists in a databse table and returns a boolean indicating existence or an error.
func EngineIDExists(engineID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM scraper_engine WHERE engine_id=?)"
	err := DB.QueryRow(query, engineID).Scan(&exists)
	if err != nil {
		InsertLog("400", "Error checking engine ID existence: "+err.Error(), "EngineIDExists()")
		return false, err
	} else {
		InsertLog("200", "Successfully checked if engine ID exists.", "EngineIDExists()")
		log.Println("Successfully checked if engine ID exists.")
	}
	return exists, nil
}

// Function to insert a new prediction
// The function InsertPrediction, that checks the existence of an engineID, logs the result, and inserts predictionInfo into a database table if the engineID exists, handling errors along the way.
func InsertPrediction(engineID string, predictionInfo string) error {
	exists, err := EngineIDExists(engineID)
	if err != nil {
		InsertLog("400", "Error checking engine ID: "+err.Error(), "InsertPrediction()")
		return fmt.Errorf("Error checking engine ID: %v", err)
	} else {
		InsertLog("200", "Successfully checked if engine ID exists.", "InsertPrediction()")
		log.Println("Successfully checked if engine ID exists.")
	}
	if !exists {
		InsertLog("400", "engine_id does not exist", "InsertPrediction()")
		return fmt.Errorf("engine_id %s does not exist", engineID)
	} else {
		InsertLog("200", "Engine ID exists.", "InsertPrediction()")
		log.Println("Engine ID exists.")
	}

	query := "INSERT INTO predictions (engine_id, prediction_info) VALUES (?, ?)"
	_, err = DB.Exec(query, engineID, predictionInfo)
	if err != nil {
		InsertLog("400", "Error storing prediction: "+err.Error(), "InsertPrediction()")
		return fmt.Errorf("Error storing prediction: %v", err)
	} else {
		InsertLog("200", "Successfully inserted prediction.", "InsertPrediction()")
		log.Println("Successfully inserted prediction.")
	}
	return nil
}

// Function to insert a sample engine ID into scraper_engine table
//
// Function inserts a sample engine's information into a database table, logs success, and returns any encountered errors.
func InsertSampleEngine(engineID, engineName, engineDescription string) error {
	query := "INSERT INTO scraper_engine (engine_id, engine_name, engine_description) VALUES (?, ?, ?)"
	_, err := DB.Exec(query, engineID, engineName, engineDescription)
	if err != nil {
		InsertLog("400", "Error inserting sample engine: "+err.Error(), "InsertSampleEngine()")
		return fmt.Errorf("Error inserting sample engine: %v", err)
	} else {
		InsertLog("200", "Successfully inserted sample engine.", "InsertSampleEngine()")
		log.Println("Successfully inserted sample engine.")
	}
	return nil
}

// Simulated ML model prediction function
//
// It definesa function that simulates an ML model prediction with a 2-second delay
// and logs a success message before returning a prediction result as a formatted string.
func PerformMLPrediction(inputData string) string {
	// Simulate some delay for ML model prediction
	time.Sleep(2 * time.Second)
	log.Println("Successfully performed ML prediction.")
	return fmt.Sprintf("Prediction result for %s", inputData)
}

// Convert prediction result to JSON
//
// defines a function that converts a given prediction result string into a JSON format, logging a success message and returning the JSON string or an error.
func ConvertPredictionToJSON(predictionResult string) (string, error) {
	predictionMap := map[string]string{"result": predictionResult}
	predictionJSON, err := json.Marshal(predictionMap)
	if err != nil {
		InsertLog("400", "Error converting prediction to JSON: "+err.Error(), "ConvertPredictionToJSON()")
		return "", err
	} else {
		InsertLog("200", "Successfully converted prediction to JSON.", "ConvertPredictionToJSON()")
		log.Println("Successfully converted prediction to JSON.")
	}
	return string(predictionJSON), nil
}
