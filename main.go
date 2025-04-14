package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

const (
	pid         = "709777522158"
	appKey      = "5003750"
	appSecret   = "lpmjpKQs6x"
	accessToken = "f537aacd-0ac4-4c8c-b6b3-d9ee63725f84"
	urlPath     = "param2/1/com.alibaba.fenxiao.crossborder/product.search.queryProductDetail/" + appKey
)

func sign(secret, urlPath string, params map[string]string) string {
	// Sort parameters by key
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build signature string
	sig := urlPath
	for _, k := range keys {
		sig += k + params[k]
	}

	// Calculate HMAC-SHA1
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(sig))
	return hex.EncodeToString(mac.Sum(nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Prepare parameters
	params := map[string]string{
		"access_token":    accessToken,
		"_aop_timestamp": strconv.FormatInt(time.Now().UnixNano()/1e6, 10),
		"offerDetailParam": `{"offerId":"` + pid + `","country":"en"}`,
	}

	// Generate signature
	signature := sign(appSecret, urlPath, params)

	// Build URL
	baseURL := "http://gw.open.1688.com/openapi/" + urlPath
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}
	query.Add("_aop_signature", signature)
	fullURL := baseURL + "?" + query.Encode()

	// Make HTTP request
	client := &http.Client{
		Timeout: time.Second * 2,
	}
	resp, err := client.Get(fullURL)
	if err != nil {
		http.Error(w, "Request failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func main() {
	http.HandleFunc("/", handler)
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	fmt.Printf("Server starting on port %s...\n", port)
	http.ListenAndServe(":"+port, nil)
}
