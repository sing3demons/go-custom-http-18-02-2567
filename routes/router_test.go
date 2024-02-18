package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	// Set the necessary environment variables
	os.Setenv("LOG_LEVEL", "debug")

	// Call the NewRouter function
	router := NewRouter()

	// Assert that the router is not nil
	assert.NotNil(t, router)

	// Perform additional assertions if needed
}

type MyContext struct {
	Response http.ResponseWriter
	Request  *http.Request
}

func TestSetParam(t *testing.T) {
	// Create a mock HTTP request
	req, _ := http.NewRequest("GET", "/hello/123", nil)
	req = setParam("/hello/{id}", req)

	paramValue := req.Context().Value(ContextKey("id"))
	expectedParamValue := "123"
	if paramValue != expectedParamValue {
		t.Errorf("Expected param value %s, got %s", expectedParamValue, paramValue)
	}
}

func TestPOST(t *testing.T) {
	// Mock path and handler function
	path := "/test"
	handler := func(ctx IContext) {
		ctx.JSON(200, "Hello")
	}

	// Create a mock HTTP request
	req, err := http.NewRequest("POST", "hello"+path, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock HTTP response recorder
	recorder := httptest.NewRecorder()
	m := &microservice{}
	m.POST(path, handler)
	http.DefaultServeMux.ServeHTTP(recorder, req)

}

func TestHTTPHandlers(t *testing.T) {
	// Initialize router
	router := NewRouter().(*microservice)

	data := struct {
		Message string `json:"message"`
	}{
		Message: "Hello, World",
	}
	// Define a test handler function
	testHandler := func(context IContext) {
		context.JSON(200, data)
	}

	// Register the test handler for a specific method and path
	router.GET("/test", testHandler)

	// Create a request to the registered endpoint
	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	http.DefaultServeMux.ServeHTTP(rr, req)

	// Check the status code is what we expect
	assert.Equal(t, http.StatusOK, rr.Code)
	expected := `{"message":"Hello, World"}` + "\n"
	assert.Equal(t, expected, rr.Body.String())

	// Perform additional assertions if needed
}

func TestPUT(t *testing.T) {
	// Mock path and handler function
	path := "/test"
	handler := func(ctx IContext) {
		ctx.JSON(200, "Hello, World")
	}

	// Create a mock HTTP request
	req, err := http.NewRequest("PUT", path, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock HTTP response recorder
	recorder := httptest.NewRecorder()
	m := &microservice{}
	m.PUT(path, handler)
	http.DefaultServeMux.ServeHTTP(recorder, req)

	// Assert the response status code
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Assert the response body
	expectedBody := `"Hello, World"`
	assert.Equal(t, expectedBody, strings.ReplaceAll(recorder.Body.String(), "\n", ""))

	// Perform additional assertions if needed
}
func TestPATCH(t *testing.T) {
	// Mock path and handler function
	path := "/test"
	handler := func(ctx IContext) {
		ctx.JSON(200, "Hello, World!")
	}

	// Create a mock HTTP request
	req, err := http.NewRequest(http.MethodPatch, path, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock HTTP response recorder
	recorder := httptest.NewRecorder()
	m := &microservice{}
	m.PATCH(path, handler)
	http.DefaultServeMux.ServeHTTP(recorder, req)

	// Assert the response status code
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Assert the response body
	expectedBody := `"Hello, World!"`
	assert.Equal(t, expectedBody, strings.ReplaceAll(recorder.Body.String(), "\n", ""))

	// Perform additional assertions if needed
}

func TestDELETE(t *testing.T) {
	// Mock path and handler function
	path := "/test"
	handler := func(ctx IContext) {
		ctx.JSON(200, "Hello, World!")
	}

	// Create a mock HTTP request
	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock HTTP response recorder
	recorder := httptest.NewRecorder()
	m := &microservice{}
	m.DELETE(path, handler)
	http.DefaultServeMux.ServeHTTP(recorder, req)

	// Assert the response status code
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Assert the response body
	expectedBody := `"Hello, World!"`
	assert.Equal(t, expectedBody, strings.ReplaceAll(recorder.Body.String(), "\n", ""))

	// Perform additional assertions if needed
}
func TestStart(t *testing.T) {
	// Set the necessary environment variables
	os.Setenv("PORT", "8080")
	os.Setenv("LOG_LEVEL", "debug") // Set log level for better visibility

	// Create a new microservice instance
	ms := NewRouter().(*microservice)

	// Call the Start method
	go func() {
		ms.Start()
	}()

	// Since Start is asynchronous, it's good to wait for a short time to ensure
	// the server has started before testing
	time.Sleep(100 * time.Millisecond)

	// Perform a request to the server to ensure it's running
	resp, err := http.Get("http://localhost:8080")
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Close the response body
	resp.Body.Close()

	// Perform additional assertions if needed
}
