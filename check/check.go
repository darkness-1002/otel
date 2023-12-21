package main

import (
	"log"
	"os"
	"runtime"

	"github.com/Jeffail/tunny"
)

func main() {
	log.SetOutput(os.Stderr)
	numCPUs := runtime.NumCPU()

	pool := tunny.NewFunc(numCPUs, func(payloadIn interface{}) interface{} {
		switch payload := payloadIn.(type) {
		case string:
			// Payload is a string
			log.Println("Received string:", payload)
			return payloadIn
		case int:
			// Payload is an integer
			payload *= 2
			log.Println("Received integer:", payload)
			return payloadIn
		case []byte:
			payloadString := string(payload)
			return payloadString
		default:
			// Payload is of some other type
			log.Printf("Received something else: %v", payload)
			return "Received an unknown type"
		}
	})
	input := "apple"

	pool.Process(input)

	// http.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
	// 	defer r.Body.Close()

	// 	input, err := ioutil.ReadAll(r.Body)
	// 	if err != nil {
	// 		http.Error(w, "Internal error", http.StatusInternalServerError)
	// 		return
	// 	}
	// 	log.Println("In handler")

	// 	// Increment the WaitGroup counter
	// 	//wg.Add(1)

	// 	//go func() {
	// 	// Decrement the WaitGroup counter when the processing is done
	// 	//	defer wg.Done()

	// 	result := pool.Process(input)

	// 	switch result := result.(type) {
	// 	case string:
	// 		w.Write([]byte(result))
	// 	default:
	// 		// Handle the case where the result is not a string
	// 		log.Printf("Unexpected result type: %v", result)
	// 		http.Error(w, "Internal error", http.StatusInternalServerError)
	// 	}
	// 	//}()
	// })

	// // Start HTTP server in a separate goroutine
	// //go func() {
	// defer pool.Close()

	// // Start the HTTP server on port 8080
	// log.Println("Server starting on :8080")
	// if err := http.ListenAndServe(":8080", nil); err != nil {
	// 	log.Fatal("Error:", err)
	// }
	// //}()

	// Wait for all HTTP requests to finish
	//select {}
}
