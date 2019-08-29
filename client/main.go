package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/jbowl/proto1/fileserv"
)

type client struct {
	fsc fileserv.FileServClient
}

func (c *client) cp(file string, dest string) error {

	name := fileserv.FileName{File: file}

	cc, err := c.fsc.GetFile(context.Background(), &name)
	if err != nil {
		log.Printf("cc.GetFile: %s", err)
		return err
	}

	var blob []byte
	for {
		c, err := cc.Recv()
		if err != nil {
			if err == io.EOF {
				log.Printf("Transfer of %d bytes successful", len(blob))
				break
			}
			log.Printf("reading from stream %s", err)
			return err
		}

		blob = append(blob, c.Chunk...)
	}

	f, err := os.Create(dest)
	if err != nil {
		log.Printf("os.Create(%s)= %s ", dest, err)
		return err
	}

	defer f.Close()

	n, err := f.Write(blob)

	if err != nil {
		log.Printf("f.Write() %s ", err)
		return err
	}
	fmt.Printf("wrote %d bytes", n)
	return nil
}

func (c *client) find(file string) (string, error) {
	name := fileserv.FileName{File: file}
	filename, err := c.fsc.FindFile(context.Background(), &name)
	if err != nil {
		return "", err
	}
	return filename.File, nil
}

func (c *client) ls(file string) error {
	name := fileserv.FileName{File: file}

	cc, err := c.fsc.LS(context.Background(), &name)
	if err != nil {
		return err
	}
	for {
		fi, err := cc.Recv()
		if err != nil {
			break
		}

		// make it look like ls -la
		t := time.Unix(fi.Unixdate, 0)
		date := fmt.Sprintf("%.3s %2d %2d:%02d", t.Month().String(), t.Day(), t.Hour(), t.Minute())
		fmt.Printf("%s %9d %s %s\n", fi.Mode, fi.Size, date, fi.Name)

	}
	return nil
}

func main() {

	filePtr := flag.String("file", "", "path to file")
	opPtr := flag.String("op", "", "choose operation")
	destPtr := flag.String("dest", "", "dest file")

	flag.Parse()

	if len(*filePtr) < 1 || len(*opPtr) < 1 {
		return // TODO: printf usage
	}

	creds, err := credentials.NewClientTLSFromFile("/home/j/certs/jsoft-server.pem", "")
	if err != nil {
		log.Printf("could not load tls cert: %s", err)
		return
	}

	conn, err := grpc.Dial("jsoft-server:8086", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Printf("grpc.Dial: %s", err)
		return
	}

	fsc := fileserv.NewFileServClient(conn)
	c := client{fsc}

	switch *opPtr {
	case "ls": // list directory
		err = c.ls(*filePtr)
	case "cp": // copy file from server
		err = c.cp(*filePtr, *destPtr)
	case "find": // look for file
		file, err := c.find(*filePtr)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("file = ", file)
	}

	if err != nil {
		fmt.Println(err)

	}
}
