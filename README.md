# Protego Development Assignment - Ron Marom

## Implementation Notes

I chose to implement the exercise in Go, using Go core library to implement the multithreading part.

With the thread pull size being an argument, for any chosen *n* sized thread pull, there will be a total of *n+3* threads running:

* **1 reader thread:** reads line by line from the input file that includes URLs and write it to a channel.

* ***n* worker threads:** responsible of getting the URL content (which is supposed to be the heaviest part of the implementation) and MD5 encoding. The worker threads read the url from the input channel and write the md5 values to the output channel. After the input channel was closed, the working threads finish processing what was in it and notify they are done via a thread group. The worker threads cannot close the output channel as they don't know when the other threads are done.

* **1 writer thread:** reads the output channel and writes the MD5 values line by line to the output files. I considered doing the writing as part of the worker thread, but that would mandate mutexing the file writing and overall not be quicker. At the end, this thread notifies it is done via a (different) wait group.

  The writer thread has a map of the order of MD5 values because of the requirement to write the MD5 values at the order of the input URLs. So, every MD5 values is stored with its original index, and then the writer thread keeps attempting to write the next index that was not written yet.

* **The main thread:** Runs all of the above threads, wait for all the worker threads to be done, closes the output channel (so the writer thread will know it is done), waits for the writer thread to signal (in order not to lose the last writings that are in progress) and exits.

## Running The Program
1. Clone this git repository
2. Go to the root folder:

  ```cd {the root path of the repository}```

3. Run:

  ```go run ./pkg/main/ -inputFile={path to the urls file} --outputFile {path to the filename the MD5 file should be created under} --threadsNum={number of worker threads}```
