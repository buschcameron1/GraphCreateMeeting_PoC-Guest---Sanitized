package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Attendee struct {
	EmailAddress struct {
		Address string `json:"address"`
		Name    string `json:"name"`
	}
	Type string `json:"type"`
}

type Event struct {
	Subject   string     `json:"subject"`
	StartTime string     `json:"start_time"`
	EndTime   string     `json:"end_time"`
	Attendees []Attendee `json:"attendees"`
	Organizer string     `json:"organizer"`
}

var apiAuth = map[string]any{
	"secret":   "[AppReg_secret]",
	"tenantID": "[Tenant ID]",
	"appID":    "[App ID]",
}

func createEventHandler(c *gin.Context) {
	var event Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	token := getBearerToken()

	eventID, eventBody := createEvent(event, token)
	if eventID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Event ID not found in response"})
		return
	}
	if eventBody == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Event body not found in response"})
		return
	}

	deleteEvent(eventID, token, event)

	body := sendInvite(event, token, eventBody)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Event created successfully",
		"body":    body,
	})
}

func deleteEvent(eventID string, token string, event Event) bool {
	URI := "https://graph.microsoft.com/v1.0/users/" + event.Organizer + "/events/" + eventID

	req, err := http.NewRequest("DELETE", URI, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return false
	}

	return true
}

func createEvent(event Event, token string) (string, string) {
	//URI := "https://graph.microsoft.com/v1.0/me/events"
	URI := "https://graph.microsoft.com/v1.0/users/" + event.Organizer + "/events"

	eventPayload := map[string]interface{}{
		"subject": event.Subject,
		"start": map[string]string{
			"dateTime": event.StartTime,
			"timeZone": "UTC",
		},
		"end": map[string]string{
			"dateTime": event.EndTime,
			"timeZone": "UTC",
		},
		"isOnlineMeeting":       true,
		"onlineMeetingProvider": "teamsForBusiness",
	}

	body, err := json.Marshal(eventPayload)
	if err != nil {
		return "", ""
	}

	req, err := http.NewRequest("POST", URI, io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return "", ""
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()

	var respBody interface{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		fmt.Print(err)
		return "", ""
	}

	eventID, ok := respBody.(map[string]interface{})["id"].(string)
	if !ok {
		return "", ""
	}

	eventBody, ok := respBody.(map[string]interface{})["body"].(map[string]interface{})["content"].(string)
	if !ok {
		return "", ""
	}

	return eventID, eventBody
}

func sendInvite(event Event, token string, body string) string {
	//URI := "https://graph.microsoft.com/v1.0/me/events"
	URI := "https://graph.microsoft.com/v1.0/users/" + event.Organizer + "/events"

	eventPayload := map[string]interface{}{
		"subject": event.Subject,
		"start": map[string]string{
			"dateTime": event.StartTime,
			"timeZone": "UTC",
		},
		"end": map[string]string{
			"dateTime": event.EndTime,
			"timeZone": "UTC",
		},
		"isOnlineMeeting": false,
		"attendees":       event.Attendees,
		"body": map[string]string{
			"contentType": "HTML",
			"content":     body,
		},
	}

	reqBody, err := json.Marshal(eventPayload)
	if err != nil {
		return "failed to marshal event payload: " + err.Error()
	}

	req, err := http.NewRequest("POST", URI, io.NopCloser(bytes.NewReader(reqBody)))
	if err != nil {
		return "failed to create HTTP request: " + err.Error()
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "failed to send request to Graph API: " + err.Error()
	}
	defer resp.Body.Close()

	var respBody interface{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		fmt.Print(err)
		return ""
	}

	respBodyArr, ok := respBody.(map[string]interface{})["body"].(map[string]interface{})["content"]
	if !ok {
		return "response body is not a valid map"
	}

	return respBodyArr.(string)
}

func getBearerToken() string {
	body := "client_id=" + apiAuth["appID"].(string) +
		"&scope=https%3A%2F%2Fgraph.microsoft.com%2F.default" +
		"&client_secret=" + apiAuth["secret"].(string) +
		"&grant_type=client_credentials"

	req, err := http.NewRequest("POST", "https://login.microsoftonline.com/"+apiAuth["tenantID"].(string)+"/oauth2/v2.0/token", bytes.NewBufferString(body))
	if err != nil {
		fmt.Println("Error creating HTTP Request, error is: ", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request to token API, error is: ", err)

	}
	defer resp.Body.Close()

	var respBody interface{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		fmt.Print(err)
		return ""
	}

	return respBody.(map[string]interface{})["access_token"].(string)
}

func main() {
	r := gin.Default()
	r.Use(cors.Default())

	r.POST("/create-event", createEventHandler)
	r.Run(":8080")
}
