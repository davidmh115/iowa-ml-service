package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/sajari/regression"
)

type TrainingRow struct {
	Year        json.Number `json:"year"`
	Month       json.Number `json:"month"`
	BottlesSold json.Number `json:"bottles_sold"`
}

type PredictRequest struct {
	Year  json.Number `json:"year"`
	Month json.Number `json:"month"`
}

type PredictResponse struct {
	Year                  float64 `json:"year"`
	Month                 float64 `json:"month"`
	PredictedBottlesSold  float64 `json:"predicted_bottles_sold"`
}

var (
	trainingData []TrainingRow
	model        regression.Regression
	modelReady   bool
)

func loadTrainingData(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.UseNumber() // important: preserves numbers as json.Number (works with "2015" or 2015)

	return dec.Decode(&trainingData)
}

func trainModel() error {
	// Simple linear regression:
	// bottles_sold = f(year, month)
	model = regression.Regression{}
	model.SetObserved("bottles_sold")
	model.SetVar(0, "year")
	model.SetVar(1, "month")

	for _, row := range trainingData {
		yearF, err := row.Year.Float64()
		if err != nil {
			return fmt.Errorf("invalid year in training data: %v", err)
		}
		monthF, err := row.Month.Float64()
		if err != nil {
			return fmt.Errorf("invalid month in training data: %v", err)
		}
		bottlesF, err := row.BottlesSold.Float64()
		if err != nil {
			return fmt.Errorf("invalid bottles_sold in training data: %v", err)
		}

		model.Train(regression.DataPoint(
			bottlesF,
			[]float64{yearF, monthF},
		))
	}

	if err := model.Run(); err != nil {
		return err
	}

	modelReady = true
	return nil
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func GetDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(trainingData)
}

func PredictHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	if !modelReady {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":"model not ready"}`))
		return
	}

	dec := json.NewDecoder(r.Body)
	dec.UseNumber() // important: also accept numbers-as-strings in requests

	var req PredictRequest
	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid JSON body"}`))
		return
	}

	yearF, err := req.Year.Float64()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"year must be a number"}`))
		return
	}
	monthF, err := req.Month.Float64()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"month must be a number"}`))
		return
	}

	pred, err := model.Predict([]float64{yearF, monthF})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"prediction failed"}`))
		return
	}

	resp := PredictResponse{
		Year:                 yearF,
		Month:                monthF,
		PredictedBottlesSold: pred,
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func main() {
	// Load JSON training data (your BigQuery export is OK even if numbers are quoted)
	if err := loadTrainingData("./data/training.json"); err != nil {
		fmt.Println("Failed to load training data:", err)
		os.Exit(1)
	}

	// Train an in-memory model using an ML library
	if err := trainModel(); err != nil {
		fmt.Println("Failed to train model:", err)
		os.Exit(1)
	}

	router := mux.NewRouter()
	router.HandleFunc("/health", HealthHandler).Methods("GET")
	router.HandleFunc("/data", GetDataHandler).Methods("GET")
	router.HandleFunc("/predict", PredictHandler).Methods("POST")

	// Cloud Run sets PORT. Default locally to 9090.
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	server := http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	fmt.Println("Starting ML microservice on port", port)
	_ = server.ListenAndServe()
}
