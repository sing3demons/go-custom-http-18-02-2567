package routes

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPContextQuery(t *testing.T) {
	// Create a mock request with query parameters
	req := httptest.NewRequest("GET", "/?name=John&age=30", nil)

	// Create an HTTPContext instance
	ctx := &HTTPContext{r: req}

	// Test Query method
	assert.Equal(t, "John", ctx.Query("name"))
	assert.Equal(t, "30", ctx.Query("age"))
	assert.Equal(t, "", ctx.Query("unknown")) // Testing for non-existing query parameter
}

func TestHTTPContextParam(t *testing.T) {
	// Create a mock request with context value
	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, ContextKey("key"), "value")
	req = req.WithContext(ctx)

	// Create a mock response writer
	w := httptest.NewRecorder()

	// Create an instance of HTTPContext
	c := &HTTPContext{w: w, r: req}

	// Test Param method
	expected := "value"
	result := c.Param("key")
	if result != expected {
		t.Errorf("Param(\"key\") returned %s, expected %s", result, expected)
	}
}

func TestHTTPContextReadBody(t *testing.T) {
	// Create a mock request with a JSON body
	body := `{"name": "John", "age": 30}`
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTPContext instance
	ctx := &HTTPContext{r: req}

	// Create a struct to unmarshal the JSON body into
	var data struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Test ReadBody method
	assert.NoError(t, ctx.Bind(&data))
	assert.Equal(t, "John", data.Name)
	assert.Equal(t, 30, data.Age)
}

func TestHTTPContextJSON(t *testing.T) {
	// Create a mock response writer
	w := httptest.NewRecorder()

	// Create an HTTPContext instance
	ctx := &HTTPContext{w: w}

	// Create a struct to encode as JSON
	data := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		Name: "John",
		Age:  30,
	}

	// Test JSON method
	ctx.JSON(http.StatusOK, data)

	// Check the response status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check the response body
	expected := `{"name":"John","age":30}` + "\n"
	assert.Equal(t, expected, w.Body.String())
}

func TestNewMyContext(t *testing.T) {
	// Create a mock HTTP request
	req, err := http.NewRequest("GET", "http://example.com/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock HTTP response recorder
	recorder := &mockResponseWriter{}

	// Call the NewMyContext function
	ctx := NewMyContext(recorder, req)

	// Check if ctx implements the IContext interface
	_, ok := ctx.(IContext)
	if !ok {
		t.Error("NewMyContext should return a value implementing the IContext interface")
	}

	// Check if ctx is of the correct underlying type (HTTPContext)
	_, ok = ctx.(*HTTPContext)
	if !ok {
		t.Error("NewMyContext should return a value of type *HTTPContext")
	}

	// You can add more assertions here to verify the values of w and r in the returned context
}

// Mock response writer for testing
type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() http.Header {
	return http.Header{}
}

func (m *mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {

}

func TestHTTPContextBind(t *testing.T) {
	// Define a test struct to bind the request body into
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Create a mock HTTP request with a JSON request body
	requestBody := []byte(`{"name": "John", "age": 30}`)
	req, err := http.NewRequest("POST", "http://example.com/test", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock HTTP context with the mock request
	ctx := &HTTPContext{w: nil, r: req}

	// Create an instance of the target object (TestStruct)
	var testObj TestStruct

	// Call the Bind method
	if err := ctx.Bind(&testObj); err != nil {
		t.Errorf("Bind returned an error: %v", err)
		return
	}

	// Verify the values of the decoded object
	expectedName := "John"
	expectedAge := 30
	if testObj.Name != expectedName || testObj.Age != expectedAge {
		t.Errorf("Unexpected values in the decoded object. Expected: %+v, Got: %+v", TestStruct{Name: expectedName, Age: expectedAge}, testObj)
	}
}
