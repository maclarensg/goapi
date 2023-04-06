package main

import (
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestIsRunningInKubernetes(t *testing.T) {
	// Create a temporary directory for the test
	dir := "/var/run/secrets/kubernetes.io"

	// Simulate k8s by creating dir
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	// Create a file to simulate the service account file
	filePath := dir + "/serviceaccount"
	if _, err := os.Create(filePath); err != nil {
		t.Fatalf("Failed to create  file: %v", err)
	}

	// Call the isRunningInKubernetes function
	result := isRunningInKubernetes()

	// Check that the result is true
	if !result {
		t.Errorf("Expected true, but got false")
	}

	// Remove the dummy file to simulate running outside of Kubernetes
	if err := os.Remove(filePath); err != nil {
		t.Fatalf("Failed to remove dummy file: %v", err)
	}

	// Call the isRunningInKubernetes function again
	result = isRunningInKubernetes()

	// Check that the result is false
	if result {
		t.Errorf("Expected false, but got true")
	}
}

func TestInitRedis(t *testing.T) {
	// Create a new mock Redis server using miniredis
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create mock Redis server: %v", err)
	}
	defer s.Close()

	// Create a new Redis client with the mock server using redismock
	initRedis("127.0.0.1", s.Server().Addr().Port, "", 0)

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Failed to ping mock Redis server: %v", err)
	}
}

func TestLogQuery(t *testing.T) {
	// Create a new mock Redis server using miniredis
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create mock Redis server: %v", err)
	}
	defer s.Close()

	// Create a new Redis client with the mock server using redismock
	initRedis("127.0.0.1", s.Server().Addr().Port, "", 0)

	// Create a new Query to log
	q := Query{
		Addresses: []Address{{IP: "127.0.0.1"}},
		ClientIP:  "127.0.0.1",
		CreatedAt: 1234567890,
		Domain:    "example.com",
	}

	logQuery(q)

}
