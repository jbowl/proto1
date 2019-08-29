package main

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/jbowl/proto1/fileserv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func find(dir string, filename string) (fqn string, err error) {
	err = filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil { // TODO : parse err correctly
				return filepath.SkipDir
			}
			//	if !info.Mode().IsRegular() {
			//			return nil
			//		}
			if info.Name() == filename {
				fqn = path
				return io.EOF
			}
			return nil
		})
	if err == io.EOF {
		err = nil
	}
	if len(fqn) == 0 {
		err = status.Error(codes.NotFound, "file not found")
	}
	return fqn, err
}

type server struct {
}

const chunkSize = 64 * 1024 // 64 KiB

// GetFile - stream file bytes to client
func (s *server) GetFile(f *fileserv.FileName, stream fileserv.FileServ_GetFileServer) error {
	filebytes, err := ioutil.ReadFile(f.File)
	if err != nil {
		return err
	}
	//	defer file.Close()

	piece := &fileserv.FilePiece{}
	for currentByte := 0; currentByte < len(filebytes); currentByte += chunkSize {
		if currentByte+chunkSize > len(filebytes) {
			piece.Chunk = filebytes[currentByte:len(filebytes)]
		} else {
			piece.Chunk = filebytes[currentByte : currentByte+chunkSize]
		}
		if err := stream.Send(piece); err != nil {
			return err
		}
	}
	return nil
}

// FindFile - look for file starting at arbitrary path and return first match with acutual path,
//      /fileToFind will  start search at /
//      /home/usr  will start search at /home/usr
func (s *server) FindFile(ctx context.Context, f *fileserv.FileName) (*fileserv.FileName, error) {

	dir, filename := filepath.Split(f.File)
	if len(dir) < 1 {
		dir = "/"
	}

	filePath, err := find(dir, filename)

	if err != nil {
		return nil, err
	}

	file := &fileserv.FileName{File: filePath}

	return file, nil
}

// LS - think ls -la ; stream to client listing of directory
func (s *server) LS(f *fileserv.FileName, stream fileserv.FileServ_LSServer) error {
	files, err := ioutil.ReadDir(f.File)
	if err != nil {
		return err
	}

	for _, file := range files {
		fi := fileserv.FileInfo{Mode: file.Mode().String(),
			Size:     file.Size(),
			Unixdate: file.ModTime().Unix(),
			Name:     file.Name()}

		err = stream.Send(&fi)
		if err != nil {
		}
	}

	return err
}

func main() {
	lis, err := net.Listen("tcp", "jsoft-server:8086")
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	creds, err := credentials.NewServerTLSFromFile("/home/j/certs/jsoft-server.pem",
		"/home/j/certs/jsoft-server-key.pem")
	if err != nil {
		log.Fatalf("could not load TLS keys: %s", err)
	}

	srv := grpc.NewServer(grpc.Creds(creds))

	fileserv.RegisterFileServServer(srv, &server{})
	srv.Serve(lis)
}
