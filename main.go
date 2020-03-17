package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	stream "github.com/GetStream/stream-chat-go/v2"
	"github.com/joho/godotenv"
)

var APIKey string
var APISecret string
var ServerSideClient stream.Client

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	APIKey = os.Getenv("STREAM_API_KEY")
	APISecret = os.Getenv("STREAM_API_SECRET")
}

func join(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		newUser := &stream.User{}

		err := json.NewDecoder(r.Body).Decode(newUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create the user or update if user already exists
		user, err := ServerSideClient.UpdateUser(newUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := ServerSideClient.CreateToken(user.ID, time.Now().Add(time.Minute*time.Duration(60)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Returns the already create General channel
		channel, err := ServerSideClient.CreateChannel("team", "general", "admin", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Add the user to the General channel
		err = channel.AddMembers([]string{user.ID}, &stream.Message{
			User: &stream.User{
				ID: user.ID,
			},
			Text: user.ID + " Joined the General channel",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			User   stream.User `json:"user"`
			Token  string      `json:"token"`
			APIKey string      `json:"api_key"`
		}{
			User:   *user,
			Token:  string(token),
			APIKey: APIKey,
		})

	default:
		fmt.Fprintf(w, "Wrong Method.")
	}
}

func main() {
	port := os.Getenv("PORT")

	client, err := stream.NewClient(APIKey, []byte(APISecret))
	if err != nil {
		log.Fatal(err)
	}

	ServerSideClient = *client

	// Create admin user or update if admin already exists
	_, err = client.UpdateUser(&stream.User{
		ID:   "admin",
		Role: "admin",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create General channel
	_, err = client.CreateChannel("team", "general", "admin", map[string]interface{}{
		"name": "General",
	})
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/join", join)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
