// main_test.go

package main

import (
    "os"
    "testing"   
    "log"

    "net/http"
    "net/http/httptest"
    "bytes"
    "encoding/json"
    "strconv"
)

var a App

func TestMain(m *testing.M) {
    a.Initialize(
		"test",
		"test",
		"test")
    ensureTableExists()
    code := m.Run()
    clearTable()
    os.Exit(code)
}

func ensureTableExists() {
    if _, err := a.DB.Exec(tableCreationQuery); err != nil {
        log.Fatal(err)
    }
}

func clearTable() {
    a.DB.Exec("DELETE FROM products")
    a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products
(
    id SERIAL,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    changed INTEGER,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`


func TestEmptyTable(t *testing.T) {
    clearTable()

    req, _ := http.NewRequest("GET", "/products", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    if body := response.Body.String(); body != "[]" {
        t.Errorf("Expected an empty array. Got %s", body)
    }
}


func executeRequest(req *http.Request) *httptest.ResponseRecorder {
    rr := httptest.NewRecorder()
    a.Router.ServeHTTP(rr, req)

    return rr
}


func checkResponseCode(t *testing.T, expected, actual int) {
    if expected != actual {
        t.Errorf("Expected response code %d. Got %d\n", expected, actual)
    }
}


func TestGetNonExistentProduct(t *testing.T) {
    clearTable()

    req, _ := http.NewRequest("GET", "/product/11", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusNotFound, response.Code)

    var m map[string]string
    json.Unmarshal(response.Body.Bytes(), &m)
    if m["error"] != "Product not found" {
        t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", m["error"])
    }
}


func TestCreateProduct(t *testing.T) {

    clearTable()

    var jsonStr = []byte(`{"name":"test product", "description":"descript", "price": 11.22, "changed":1}`)
    req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")

    response := executeRequest(req)
    checkResponseCode(t, http.StatusCreated, response.Code)

    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)

    if m["name"] != "test product" {
        t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
    }

    if m["price"] != 11.22 {
        t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
    }

    // the id is compared to 1.0 because JSON unmarshaling converts numbers to
    // floats, when the target is a map[string]interface{}
    if m["id"] != 1.0 {
        t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
    }
}


func TestGetProduct(t *testing.T) {
    clearTable()
    addProducts(1)

    req, _ := http.NewRequest("GET", "/product/1", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)
}

func TestGetUpdatedProducts(t *testing.T) {
    clearTable()
    addProducts(5)

    req, _ := http.NewRequest("GET", "/products/3", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)
    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)

}


func addProducts(count int) {
    if count < 1 {
        count = 1
    }

    for i := 0; i < count; i++ {
        a.DB.Exec("INSERT INTO products(name, description, price, changed) VALUES($1, $2, $3, $4)", "Product "+strconv.Itoa(i),"", (i+1.0)*10, 5)
    }
}


func TestDeleteProduct(t *testing.T) {
    clearTable()
    addProducts(1)

    req, _ := http.NewRequest("GET", "/product/1", nil)
    response := executeRequest(req)
    checkResponseCode(t, http.StatusOK, response.Code)

    req, _ = http.NewRequest("DELETE", "/product/1", nil)
    response = executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    req, _ = http.
    
    NewRequest("GET", "/product/1", nil)
    response = executeRequest(req)
    checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestUpdateProduct(t *testing.T) {

    clearTable()
    addProducts(1)

    req, _ := http.NewRequest("GET", "/product/1", nil)
    response := executeRequest(req)
    var originalProduct map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &originalProduct)

    var jsonStr = []byte(`{"name":"test product - updated name", "description":"updated descript", "price": 11.22, "changed": 999}`)
    req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")

    response = executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    // req, _ := http.NewRequest("GET", "/product/1", nil)
    // response := executeRequest(req)

    // checkResponseCode(t, http.StatusOK, response.Code)


    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)

    if m["id"] != originalProduct["id"] {
        t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
    }

    if m["name"] == originalProduct["name"] {
        t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
    }

    if m["price"] == originalProduct["price"] {
        t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
    }
    if m["changed"] == originalProduct["changed"] {
        t.Errorf("Expected the changed to change from '%v' to '%v'. Got '%v'", originalProduct["changed"], m["changed"], m["changed"])
    }
    if m["description"] == originalProduct["description"] {
        t.Errorf("Expected the changed to change from '%v' to '%v'. Got '%v'", originalProduct["description"], m["description"], m["description"])
    }
}

func TestUpdateProductOutdated(t *testing.T) {

    clearTable()
    addProducts(1)

    req, _ := http.NewRequest("GET", "/product/1", nil)
    response := executeRequest(req)
    var originalProduct map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &originalProduct)

    var jsonStr = []byte(`{"name":"test product - updated name", "description":"updated descript", "price": 11.22, "changed": 1}`)
    req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")

    response = executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)

    if m["id"] != originalProduct["id"] {
        t.Errorf("Expected the id not to remain the same (%v). Got %v", originalProduct["id"], m["id"])
    }

    if m["name"] != originalProduct["name"] {
        t.Errorf("Expected the name not to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
    }

    if m["price"] != originalProduct["price"] {
        t.Errorf("Expected the price not to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
    }

    if m["changed"] != originalProduct["changed"] {
        t.Errorf("Expected the changed not to change from '%v' to '%v'. Got '%v'", originalProduct["changed"], m["changed"], m["changed"])
    }
    if m["description"] != originalProduct["description"] {
        t.Errorf("Expected the description not to change from '%v' to '%v'. Got '%v'", originalProduct["description"], m["description"], m["description"])
    }
}
