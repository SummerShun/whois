package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// 查询 whois 信息
func whoisQuery(server, query string) (string, error) {
	conn, err := net.Dial("tcp", server+":43")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(query + "\r\n"))
	if err != nil {
		return "", err
	}

	result, err := ioutil.ReadAll(conn)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// 获取顶级域的 whois 服务器
func getWhoisServer(domain string) (string, error) {
	result, err := whoisQuery("whois.iana.org", domain)
	if err != nil {
		return "", err
	}

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "whois:") {
			return strings.TrimSpace(strings.Split(line, ":")[1]), nil
		}
	}

	return "", fmt.Errorf("whois server not found")
}

// 处理 whois 查询请求
func handleWhoisQuery(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "Missing 'domain' query parameter", http.StatusBadRequest)
		log.Printf("Missing 'domain' query parameter")
		return
	}

	tld := domain[strings.LastIndex(domain, ".")+1:]
	whoisServer, err := getWhoisServer(tld)
	if err != nil {
		http.Error(w, "Error getting whois server: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error getting whois server for domain %s: %v", domain, err)
		return
	}

	whoisInfo, err := whoisQuery(whoisServer, domain)
	if err != nil {
		http.Error(w, "Error querying whois information: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error querying whois information for domain %s: %v", domain, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(whoisInfo))

	duration := time.Since(start)
	log.Printf("Served whois request for domain %s in %v", domain, duration)
}

func main() {
	port := flag.String("port", "8080", "Port to run the HTTP server on")
	flag.Parse()

	http.HandleFunc("/whois", handleWhoisQuery)

	fmt.Println("Starting server on port " + *port)
	log.Printf("Starting server on port %s", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

