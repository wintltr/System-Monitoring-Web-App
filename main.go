package main

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/goroutines"
	"github.com/wintltr/login-api/routes"
)

func main() {
	go goroutines.CheckClientOnlineStatusGour()
	router := mux.NewRouter().StrictSlash(true)
	credentials := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})
	// Login
	router.HandleFunc("/login", routes.Login).Methods("POST", "OPTIONS")

	// SSH Connection
	router.HandleFunc("/sshconnection", routes.SSHCopyKey).Methods("POST", "OPTIONS")
	router.HandleFunc("/sshconnection/{id}/test", routes.TestSSHConnection).Methods("GET", "OPTIONS")
	router.HandleFunc("/sshconnections", routes.GetAllSSHConnection).Methods("GET", "OPTIONS")
	router.HandleFunc("/sshconnection/{id}", routes.SSHConnectionDeleteRoute).Methods("DELETE", "OPTIONS")

	// SSH Key
	router.HandleFunc("/sshkey", routes.AddSSHKey).Methods("POST", "OPTIONS")
	router.HandleFunc("/sshkey/{id}", routes.SSHKeyDeleteRoute).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/sshkeys", routes.GetAllSSHKey).Methods("GET", "OPTIONS")

	// Get PC info
	router.HandleFunc("/systeminfo/{id}", routes.GetSystemInfoRoute).Methods("GET", "OPTIONS")
	router.HandleFunc("/systeminfos", routes.SystemInfoGetAllRoute).Methods("GET", "OPTIONS")
	router.HandleFunc("/receivelog", routes.Receivelog).Methods("POST", "OPTIONS")

	// Network Function
	router.HandleFunc("/network/defaultip", routes.GetAllDefaultIP).Methods("GET")

	// Package
	router.HandleFunc("/package/install", routes.PackageInstall).Methods("POST")
	router.HandleFunc("/package/remove", routes.PackageRemove).Methods("POST")
	router.HandleFunc("/package/list", routes.PackageListAll).Methods("POST")

	// Host User
	router.HandleFunc("/hostuser/add", routes.HostUserAdd).Methods("POST")
	router.HandleFunc("/hostuser/remove", routes.HostUserRemove).Methods("POST")
	router.HandleFunc("/hostuser/list/{id}", routes.HostUserListAll).Methods("GET")

	// User command history
	router.HandleFunc("/history/list/{id}", routes.HistoryListAll).Methods("GET")

	// Event Web
	router.HandleFunc("/eventweb", routes.GetAllEventWeb).Methods("GET")
	router.HandleFunc("/eventweb/delete/all", routes.DeleteAllEventWeb).Methods("GET")

	// Custom API
	router.HandleFunc("/pcs", routes.GetAllPcs).Methods("GET")

	log.Fatal(http.ListenAndServe(":8081", handlers.CORS(credentials, methods, origins)(router)))
}
