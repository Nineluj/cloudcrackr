package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

func getLines(filePath string) []string {
	file, _ := os.Open(filePath)
	defer file.Close()
	fileBytes, _ := ioutil.ReadAll(file)
	return strings.Split(string(fileBytes), "\n")
}

func userPassCracker(username, hash string, dict []string, outputFile *os.File,
	fileLock *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	for _, word := range dict {
		h := md5.New()
		h.Write([]byte(word))
		result := string(h.Sum(nil))

		if result == hash {
			fileLock.Lock()
			defer fileLock.Unlock()

			_, _ = outputFile.Write([]byte(fmt.Sprintf("%v:%v:%v\n", username, hash, word)))
			return
		}
	}
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: [hashes] [dictionary] [output]")
	}

	// Get the required data from the hash and dictionary files
	hashLines := getLines(os.Args[1])

	users := make(map[string]string)
	for _, line := range hashLines {
		parts := strings.SplitN(line, ":", 2)
		users[parts[0]] = parts[1]
	}

	dict := getLines(os.Args[2])
	mutex := sync.Mutex{}

	// Open the output file for writing
	outFile, err := os.OpenFile(os.Args[3], os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	var wg sync.WaitGroup

	for username, hash := range users {
		go userPassCracker(username, hash, dict, outFile, &mutex, &wg)
		wg.Add(1)
	}

	wg.Wait()
}
