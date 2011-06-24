package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"chunkymonkey/record"
	parser "util/intercept/intercept_parse"
)

var recordBase = flag.String(
	"record", "", "Record player connections to files with this prefix")

type RelayReport struct {
	written int64
	err     os.Error
}

// Sends a report on reportChan when it completes
func spliceParser(parser func(reader io.Reader), dst io.Writer, src io.Reader) (reportChan chan RelayReport) {

	parserReader, parserWriter := io.Pipe()
	wrappedDst := io.MultiWriter(dst, parserWriter)
	reportChan = make(chan RelayReport)

	go parser(parserReader)
	go func() {
		written, err := io.Copy(wrappedDst, src)
		reportChan <- RelayReport{written, err}
	}()

	return
}

func serveConn(clientConn net.Conn, remoteaddr string, connNumber int) {
	defer clientConn.Close()

	clientAddr := clientConn.RemoteAddr().String()
	log.Printf("[%d](%s) Client connected", connNumber, clientAddr)

	log.Printf("[%d](%s) Creating relay to server %s", connNumber, clientAddr, remoteaddr)
	serverConn, err := net.Dial("tcp", remoteaddr)
	if err != nil {
		log.Printf("[%d](%s) Error connecting to %s: %v", connNumber, clientAddr, remoteaddr, err)
		return
	}
	defer serverConn.Close()
	log.Printf("[%d](%s) Connected to server %s", connNumber, clientAddr, remoteaddr)

	// clientReader reads data sent from the client.
	clientReader := io.Reader(clientConn)
	if *recordBase != "" {
		recordFilename := fmt.Sprintf("%s-%d.record", *recordBase, connNumber)
		recordOutput, err := os.Create(recordFilename)
		if err != nil {
			log.Printf(
				"[%d](%s) Failed to open file %q to record connection: %v",
				connNumber, clientAddr, recordFilename, err)
		} else {
			recorder := record.NewReaderRecorder(recordOutput, clientConn)
			defer recorder.Close()
			clientReader = recorder
		}
	}

	clientParser := new(parser.MessageParser)
	serverParser := new(parser.MessageParser)

	// Set up for parsing messages from server to client
	serverToClientReportChan := spliceParser(
		func(reader io.Reader) { serverParser.ScParse(reader, connNumber) },
		clientConn, serverConn)

	// Set up for parsing messages from client to server
	clientToServerReportChan := spliceParser(
		func(reader io.Reader) { clientParser.CsParse(reader, connNumber) },
		serverConn, clientReader)

	// Wait for the both relay/splices to stop, then we let the connections
	// close via deferred calls
	report := <-serverToClientReportChan
	log.Printf("[%d](%s) Server->client relay after %d bytes with error: %v", connNumber, clientAddr, report.written, report.err)
	report = <-clientToServerReportChan
	log.Printf("[%d](%s) Client->server relay after %d bytes with error: %v", connNumber, clientAddr, report.written, report.err)

	log.Printf("[%d](%s) Client disconnected", connNumber, clientAddr)
}

func serve(localaddr, remoteaddr string) (err os.Error) {
	listener, err := net.Listen("tcp", localaddr)
	if err != nil {
		log.Fatal("Listen: ", err.String())
		return
	}

	defer listener.Close()
	log.Print("Listening on ", localaddr)

	connNumber := 1

	for {
		clientConn, acceptErr := listener.Accept()
		if acceptErr != nil {
			log.Printf("Accept error: %s", acceptErr.String())
			break
		} else {
			go serveConn(clientConn, remoteaddr, connNumber)
			connNumber++
		}
	}
	return
}

func usage() {
	os.Stderr.WriteString("usage: " + os.Args[0] + " [options] localaddr:port remoteaddr:port\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	localaddr := flag.Arg(0)
	remoteaddr := flag.Arg(1)

	// It's nice to have high time precision when looking at packets
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	serve(localaddr, remoteaddr)
}
