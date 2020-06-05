package main

import (
	"log"
	"net/http"

	"os/exec"
	"encoding/json"
	"strings"
	"bytes"
)

// Loop through out hosts and collect information
func main() {
	hostnames := []string{ "hnbe1", "hnbe2", "hnbe3", "hnbe4", "hcmbe1", "hcmbe2", "hcmbe3", "hcmbe4" }
	for _, s := range hostnames {
	    doCheckStatus(s)
	}
}

// Perform file check status for our app
func doCheckStatus(host string) {
	endpoint := "http://localhost:18844/msg"
	remoteHost := host + "smsct.vnpt.vn"
	remotePath := ":/tango/config/*"
	localPath := "/tango/data/config/live/" + host
	rsyncCmd := exec.Command("/bin/rsync", "-a", "--delete", "--exclude", ".*", remoteHost + remotePath, localPath + "/")
	rsyncOut, err := rsyncCmd.Output()
	if err != nil {
		log.Printf("Error running rsync: %s", err)
		return
	}
	log.Printf("1 - Sync data from %s for directory %s: %s", remoteHost, localPath, rsyncOut)

	// git fetch origin master
	// git reset --hard origin/master
	gitFetchCmd := exec.Command("/usr/bin/git", "fetch")
	gitFetchCmd.Dir = localPath
	gitFetchOut, err := gitFetchCmd.Output()
	if err != nil {
		log.Printf("Error running git fetch: %s", err)
		return
	}

	gitResetCmd := exec.Command("/usr/bin/git", "reset", "origin/master")
	gitResetCmd.Dir = localPath
	gitResetOut, err := gitResetCmd.Output()
	if err != nil {
		log.Printf("Error running git reset hard: %s", err)
		return
	}
	log.Printf("2 - pull data for localPath %s: %s - %s", localPath, gitFetchOut, gitResetOut)

	log.Printf("3 - Checking status %s for host %s then send to %s", localPath, host, endpoint)
	gitStatusCmd := exec.Command("/usr/bin/git", "status", "-s")
	gitStatusCmd.Dir = localPath
	out, err := gitStatusCmd.Output()
	if err != nil {
		log.Printf("Error running git status: %s", err)
		return
	}
	status := strings.Trim(string(out), "\n")
	if len(status) > 0 {
		log.Printf("status: %s", out)

		// send the notification to OAM web socket
		message := map[string]interface{}{
			"remoteHost":  host,
			"remotePath": remotePath,
			"localPath": localPath,
			"sync": rsyncOut,
			"fetch": gitFetchOut,
			"reset": gitResetOut,
			"status": status,
		}
		bytesRepresentation, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error occur: %s", err)
			return
		}

		// Send message to web socket for showing the status
		http.Post(endpoint, "application/json", bytes.NewBuffer(bytesRepresentation))
	}
}
