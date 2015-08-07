package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

var bambooAPI string
var bambooBuildPlan string
var bambooUser string
var bambooPassword string
var bambooArtifactName string
var bambooYML string

const (
	unzipFolder = "./unzip"
)

func downloadFile(request *url.URL) (file *os.File, err error) {
	// Parse file name
	splits := strings.Split(request.Path, "/")
	fileName := splits[len(splits)-1]
	fmt.Printf("Start downloading file %v from %v.\n", fileName, request)

	// Create local file handler
	output, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	defer output.Close()
	fmt.Printf("Local file handler to %v created.\n", fileName)

	// Make file request
	client := &http.Client{}
	req, err := http.NewRequest("GET", request.String(), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(bambooUser, bambooPassword)
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	fmt.Println("GET request to Bamboo for file made.")

	// Copy request data into local file
	n, err := io.Copy(output, response.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Download of %v (%v Bytes) complete.\n %v", fileName, n)
	return output, nil
}

func getLatestArtifact() (url *url.URL, err error) {
	concat := fmt.Sprintf("%v/result/%v/latest.json?expand=artifacts", bambooAPI, bambooBuildPlan)

	// Make API request
	client := &http.Client{}
	req, err := http.NewRequest("GET", concat, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(bambooUser, bambooPassword)
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	fmt.Println("GET request to Bamboo for plan details made.")

	// Parse JSON
	dec := json.NewDecoder(response.Body)
	var v map[string]interface{}
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	artifacts := v["artifacts"].(map[string]interface{})
	artifactsArray := artifacts["artifact"].([]interface{})
	var artifactDetails map[string]interface{}
	var artifactLink map[string]interface{}

	for _, value := range artifactsArray {
		artifactDetails = value.(map[string]interface{})
		if artifactDetails["name"] == bambooArtifactName {
			fmt.Printf("Artifact \"%v\" found.\n", bambooArtifactName)
			artifactLink = artifactDetails["link"].(map[string]interface{})
		}
	}

	// Parse Artifact URL
	url, err = url.Parse(artifactLink["href"].(string))
	if err != nil {
		return nil, err
	}
	fmt.Printf("Link to artifact is %v.\n", url)
	return url, nil
}

func unzipFile(file, destination string) error {
	unzip := exec.Command("unzip", "-o", file, "-d", destination)
	fmt.Printf("Unzip %v to %v.\n", file, destination)
	unzip.Stdout = os.Stdout
	unzip.Stderr = os.Stderr
	return unzip.Run()
}

func runComposition(project, yml string) error {
	build := exec.Command("docker-compose", "-p", project, "-f", yml, "build")
	fmt.Println("Build composition.")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	err := build.Run()
	if err != nil {
		return err
	}

	compose := exec.Command("docker-compose", "-p", project, "-f", yml, "up", "-d")
	fmt.Println("Start composition.")
	compose.Stdout = os.Stdout
	compose.Stderr = os.Stderr
	return compose.Run()
}

func main() {

	// Get environment variables
	bambooAPI = os.Getenv("bambooAPI")
	bambooBuildPlan = os.Getenv("bambooBuildPlan")
	bambooUser = os.Getenv("bambooUser")
	bambooPassword = os.Getenv("bambooPassword")
	bambooArtifactName = os.Getenv("bambooArtifactName")
	bambooYML = os.Getenv("bambooYML")
	if bambooAPI == "" || bambooBuildPlan == "" || bambooUser == "" || bambooPassword == "" || bambooArtifactName == "" || bambooYML == "" {
		fmt.Fprintf(os.Stderr, "Not all environment variables set.\n")
		os.Exit(1)
	}

	// Get latest artifact
	url, err := getLatestArtifact()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Download artifact
	file, err := downloadFile(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Unzip artifact
	err = unzipFile(file.Name(), unzipFolder)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Start composition
	err = runComposition("asy-deploy", fmt.Sprintf("%v/%v", unzipFolder, bambooYML))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Clear Up
	err = os.Remove(file.Name())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	err = os.RemoveAll(unzipFolder)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}
