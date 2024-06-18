package main

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	au "github.com/yerlans/us-protos/gen/auth-service"
	us "github.com/yerlans/us-protos/gen/us-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type APIGateway struct {
	urlShortenerClient us.UrlShorteningServiceClient
	authClient         au.AuthServiceClient
}

func NewAPIGateway(authConn *grpc.ClientConn) *APIGateway {
	return &APIGateway{
		authClient: au.NewAuthServiceClient(authConn),
	}
}

type CreateShortUrlRequest struct {
	OriginalUrl string `json:"original_url"`
}

func (a *APIGateway) CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	var req CreateShortUrlRequest
	// Parse request body and map to req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	grpcReq := &us.ShortenUrlRequest{OriginalUrl: req.OriginalUrl}

	grpcResp, err := a.urlShortenerClient.ShortenUrl(context.Background(), grpcReq)
	if err != nil {
		grpcError, _ := status.FromError(err)
		http.Error(w, grpcError.Message(), 500)
		return
	}
	// Write response to HTTP
	json.NewEncoder(w).Encode(grpcResp)
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *APIGateway) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	// Parse request body and map to req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	grpcReq := &au.RegisterRequest{Email: req.Email, Password: req.Password}

	grpcResp, err := a.authClient.Register(context.Background(), grpcReq)
	if err != nil {
		grpcError, _ := status.FromError(err)
		http.Error(w, grpcError.Message(), 500)
		return
	}
	// Write response to HTTP
	json.NewEncoder(w).Encode(grpcResp)
}

func main() {
	// Connect to gRPC service
	conn, err := grpc.NewClient("localhost:44044", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to gRPC service: %v", err)
	}
	defer conn.Close()

	apiGateway := NewAPIGateway(conn)

	r := mux.NewRouter()
	//r.HandleFunc("/shorten", apiGateway.CreateShortUrl).Methods("POST")
	r.HandleFunc("/register", apiGateway.Register).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", r))
}
