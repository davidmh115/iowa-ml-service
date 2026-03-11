package main

import "testing"

func TestTrainingDataFilePath(t *testing.T) {
	err := loadTrainingData("./data/training.json")
	if err != nil {
		t.Fatalf("expected training data to load, got error: %v", err)
	}
	if len(trainingData) == 0 {
		t.Fatal("expected training data to contain rows")
	}
}
