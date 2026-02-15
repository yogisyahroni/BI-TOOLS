package integration

const baseURL = "http://localhost:8080/api"

func getTestDSN() string {
	// Return hardcoded for now as per environment
	return "postgresql://postgres:1234@localhost:5432/Inside_engineer1?sslmode=disable"
}
