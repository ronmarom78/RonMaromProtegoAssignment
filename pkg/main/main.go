package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

type indexedText struct {
	index int
	text  string
}

func calculateMd5FromUrl(url string) (string, error) {
	httpClient := http.DefaultClient
	response, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("failed to read response", url, "error:", err)
		return "", err
	}
	hash := md5.Sum(responseBytes)
	result := hex.EncodeToString(hash[:])
	return result, nil
}

func worker(threadNum int, inputChannel chan indexedText, outputChannel chan indexedText, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range inputChannel {
		log.Println("Thread number ", threadNum, " is starting work on url: ", url)
		result, err := calculateMd5FromUrl(url.text)
		if err != nil {
			log.Println("failed to get url", "error:", err)
		} else {
			log.Println("Thread number ", threadNum, " finished work on url: ", url, " with result ", result)
		}
		outputChannel <- indexedText{
			index: url.index,
			text:  result,
		}
	}
}

func inputFileReader(inputFile string, inputChannel chan indexedText) {
	readFile, err := os.Open(inputFile)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)
	lineNum := 0
	for fileScanner.Scan() {
		inputChannel <- indexedText{
			index: lineNum,
			text:  fileScanner.Text(),
		}
		lineNum++
	}
	close(inputChannel)
}

func outputFileWriter(outputFile string, outputChannel chan indexedText, writing *sync.WaitGroup) {
	writeFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer writeFile.Close()
	resultsByIndex := make(map[int]string)
	nextResultToWrite := 0
	for result := range outputChannel {
		log.Println("Received result: ", result)
		resultsByIndex[result.index] = result.text
		for {
			nextResult, ok := resultsByIndex[nextResultToWrite]
			if !ok {
				break
			}
			_, err := fmt.Fprintln(writeFile, nextResult)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			nextResultToWrite++
		}
	}
	fmt.Println("file written successfully")
	writing.Done()
}

func processUrlsFile(numWorkers int, inputFile string, outputFile string) {
	log.Println("number of worker threads:", numWorkers)
	inputChannel := make(chan indexedText)
	outputChannel := make(chan indexedText)

	var processing sync.WaitGroup
	processing.Add(numWorkers)
	var writing sync.WaitGroup
	writing.Add(1)

	for w := 1; w <= numWorkers; w++ {
		go worker(w, inputChannel, outputChannel, &processing)
	}

	go inputFileReader(inputFile, inputChannel)

	go outputFileWriter(outputFile, outputChannel, &writing)

	processing.Wait()
	close(outputChannel)
	log.Println("Finished working")
	writing.Wait()
	log.Println("Finished writing")
}

func main() {
	const defaultInputFile = "./input/urls.txt"
	const defaultOutputFile = "./output/md5.txt"
	const defaultNumWorkers = 2

	inputFilePtr := flag.String("inputFile", defaultInputFile, "file to read urls from")
	outputFilePtr := flag.String("outputFile", defaultOutputFile, "file to write md5 to")
	numWorkersPtr := flag.Int("numWorkers", defaultNumWorkers, "number of processing threads")

	flag.Parse()

	processUrlsFile(*numWorkersPtr, *inputFilePtr, *outputFilePtr)
}
