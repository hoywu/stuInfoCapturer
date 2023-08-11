package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stuInfoCapturer/auth"
	"stuInfoCapturer/score"
)

func main() {
	http.HandleFunc("/GetToken", getTokenHandler)
	http.HandleFunc("/GetQRCode", getQRCodeHandler)
	http.HandleFunc("/CheckQRStatus", checkQRStatusHandler)
	http.HandleFunc("/Logged", loggedHandler)

	fmt.Println("Server started at :57314")
	http.ListenAndServe(":57314", addCorsHeaders(http.DefaultServeMux))
}

// 添加CORS中间件
func addCorsHeaders(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // 允许所有域名访问，生产环境中应限制
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	}
}

func getTokenHandler(w http.ResponseWriter, r *http.Request) {
	session, err := auth.NewSession()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v", err)
		return
	}

	jsonData, _ := json.Marshal(session)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func getQRCodeHandler(w http.ResponseWriter, r *http.Request) {
	var session auth.Session
	err := json.NewDecoder(r.Body).Decode(&session)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%v", err)
		return
	}
	imageData, err := session.GetQRCode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v", err)
		return
	}
	w.Header().Set("Content-Type", "image/png;charset=UTF-8")
	w.Write(imageData)
}

func checkQRStatusHandler(w http.ResponseWriter, r *http.Request) {
	var session auth.Session
	err := json.NewDecoder(r.Body).Decode(&session)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%v", err)
		return
	}
	status, _ := session.CheckQRStatus()
	fmt.Fprintf(w, "%d", status)
}

func loggedHandler(w http.ResponseWriter, r *http.Request) {
	var session auth.Session
	err := json.NewDecoder(r.Body).Decode(&session)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%v", err)
		return
	}

	cookie, err := session.Login()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error when session login: %v", err)
		return
	}

	zcScore, err := score.GenerateScoreFile(cookie)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error when download score: %v", err)
		return
	}

	jsonData, _ := json.Marshal(zcScore)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
