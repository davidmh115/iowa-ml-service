# Iowa Liquor Sales ML Service

A containerized Go microservice application that uses a machine learning library and is based on a small sample of my project data. Accepts JSON input and Returns JSON output. This application will predict the numnber of bottles sold based on the inputted month. 


## JSON Training Data
### I ran the following Query based on Iowa Liquor Sales Biq Query Dataset to get bottles sold each month from 2015 to 2025  

SELECT
  EXTRACT(YEAR  FROM date) AS year,
  EXTRACT(MONTH FROM date) AS month,
  SUM(bottles_sold)        AS bottles_sold
FROM `bigquery-public-data.iowa_liquor_sales.sales`
WHERE date >= '2015-01-01'
  AND date <  '2025-01-01'
GROUP BY year, month
ORDER BY year, month;


## Features

- Loads training data from `data/training.json`
- Trains a regression model at startup
- Exposes REST endpoints:
  - `GET /health`
  - `GET /data`
  - `POST /predict`

## Run locally:

go run main.go


## Once running open new terminal and run: 

curl http://localhost:9090/health

curl -X POST http://localhost:9090/predict \
  -H "Content-Type: application/json" \
  -d '{"year":"2023","month":"6"}'

# Run with Docker

## Build 

docker build -t iowa-ml-service:latest .

## Run 

docker run --rm -p 9090:9090 iowa-ml-service:latest




### Example Request 

JSON 
{
  "year": 2023,
  "month": 6,
  "predicted_bottles_sold": 2621439.2621603045
}

### Example Output 

JSON 
{
  "year": 2023,
  "month": 6,
  "predicted_bottles_sold": 2621439.2621603045
}


# Make a Test 


nano main_test.go


Paste the following: 


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






### Screenshots Demonstrating Operation

![Contents](images/screenshot1.png)

![Running Application](images/screenshot2.png)

![Predicting Bottles Sold](images/screenshot3.png)



## Monitoring with Prometheus

This project includes monitoring using **Prometheus**. 

### 1. Start the ML Service

Run the application:

```bash
go run main.go
```

The service will start on:

```
http://localhost:9090
```

You can verify the service is running by visiting the health endpoint:

```
http://localhost:9090/health
```

---

### 2. Start Prometheus

Run the Prometheus container in a new terminal:

```bash
docker run -d --name prometheus \
  -p 9091:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus
```

Verify the container is running:

```bash
docker ps
```

---

### 3. Open the Prometheus Web Interface

Open a browser and navigate to:

```
http://<VM_IP>:9091
```

(or `http://localhost:9091` if running locally)

Then navigate to:

```
Status → Targets
```

You should see the **iowa-ml-service** target listed as **UP**, indicating Prometheus is successfully scraping metrics from the application.

---

### 4. Query Metrics

Go to the **Graph** tab and run the query:

```
go_goroutines
```

This metric shows the number of active goroutines in the Go application and confirms that Prometheus is successfully collecting metrics.

---

### 5. Generate Sample Traffic

To generate traffic and populate metrics, run:

```bash
for i in {1..50}; do curl -s http://localhost:9090/health > /dev/null; done
```

This sends multiple requests to the service so Prometheus can collect metrics related to request handling and runtime behavior.




