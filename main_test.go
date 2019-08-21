package main

import (
	"testing"
)

//https://lanre.wtf/blog/2017/04/08/testing-http-handlers-go/
func TestApi(t *testing.T) {
	t.Run("test / ", func(t *testing.T) {
		// req, err := http.NewRequest("GET", "/", nil)
		// checkErr(err, t)

		// res := httptest.NewRecorder()
		// http.HandlerFunc(home).ServeHTTP(res, req)
		// if status := res.Code; status != http.StatusOK {
		// 	t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
		// }
		// assert.Equal(t, "welcome user", res.Body.String())
	})
}

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}
}
