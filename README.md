# Twitter-Client-Server-Model
A client-server model for a twitter-like application. 

##Description

The project simulates a client-server model for a simple twitter-like application. The client may send multiple requests to the server, and the server will respond the client with results after finishing processing the requests. The server has a sequential version as well as a parallel version for handling these requests. 


A thread-safe linked list data structure and a tweet feed task queue are implemented.


Requests (i.e., tasks in our program) are sent from a “client” (e.g., a redirected file on the command line, a task generator program piped into your program, etc.) via os.Stdin. The “server” (i.e., your program) will process these requests and send their results back to the client via os.Stdout.


##Program Usage


The program has the following usage and required command-line argument:


Usage: twitter <number of goroutines> <block size>

<number of goroutines> = the number of goroutines to be part of the queue

<block size> = the maximum number of tasks a goroutine can process at any given point in time.
