package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"log"
	storagev1 "tages-go/api/storage/storage"
	"tages-go/internal/service"
)

type serverAPI struct {
	maxConcurrentRead           int
	maxConcurrentUploadDownload int

	read           chan struct{}
	uploadDownload chan struct{}
	service        service.Service
	storagev1.UnimplementedStorageServer
}

func Register(gRPC *grpc.Server, service service.Service, maxRead, maxUpDownLoad int) {
	storagev1.RegisterStorageServer(gRPC, &serverAPI{
		service:        service,
		read:           make(chan struct{}, maxRead),
		uploadDownload: make(chan struct{}, maxUpDownLoad),
	})
}

func (s *serverAPI) UploadFile(stream storagev1.Storage_UploadFileServer) error {
	select {
	case s.uploadDownload <- struct{}{}:
		log.Println("New connection.")
		defer func() {
			log.Println("Connection freed.")
			<-s.uploadDownload
		}()

		var name string

		fileData := make([]byte, 0, 1024)

		for {
			req, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return status.Error(codes.InvalidArgument, "couldn't receive request")
			}

			name = req.GetName()

			if name == "" {
				return status.Error(codes.InvalidArgument, "filename is empty")
			}

			chunk := req.GetImChunk()

			fileData = append(fileData, chunk...)
		}

		if err := s.service.FileStorage.UploadFileChunk(name, fileData); err != nil {
			log.Println(err)
			return status.Error(codes.Internal, "couldn't write data to a file")
		}

		log.Printf("Received file: %s", name)

		return stream.SendAndClose(&storagev1.UploadFileResponse{Name: name})
	default:
		return status.Error(codes.PermissionDenied, "network is busy")
	}

}

func (s *serverAPI) DownloadFile(req *storagev1.FileMetadataRequest, stream storagev1.Storage_DownloadFileServer) error {
	select {
	case s.uploadDownload <- struct{}{}:
		log.Println("New connection.")
		defer func() {
			log.Println("Connection freed.")
			<-s.uploadDownload
		}()

		name := req.GetName()
		if name == "" {
			return status.Error(codes.InvalidArgument, "filename is empty")
		}

		bytes, err := s.service.FileStorage.ReadFile(name)
		if err != nil {
			log.Println(err)
			return status.Error(codes.Internal, "couldn't read file")
		}

		length := len(bytes)

		for {
			if length >= 1024 {
				if err := stream.Send(&storagev1.FileResponse{Chunk: bytes[:1024]}); err != nil {
					return status.Error(codes.Internal, "couldn't send data chunk")
				}

				bytes = bytes[1024:]
				length -= 1024
			} else {
				if err := stream.Send(&storagev1.FileResponse{Chunk: bytes}); err != nil {
					return status.Error(codes.Internal, "couldn't send data")
				}

				break
			}
		}
	default:
		return status.Error(codes.PermissionDenied, "network is busy")
	}

	return nil
}

func (s *serverAPI) FileData(ctx context.Context, empty *emptypb.Empty) (*storagev1.FileDataResponse, error) {
	select {
	case s.read <- struct{}{}:
		log.Println("New connection.")
		defer func() {
			log.Println("Connection freed.")
			<-s.read
		}()

		imData, err := s.service.FileStorage.FilesData()
		if err != nil {
			log.Println(err)
			return nil, status.Error(codes.Internal, "couldn't get files' data")
		}

		if len(imData) == 0 {
			return nil, status.Error(codes.NotFound, "no files in storage")
		}

		resp := make([]*storagev1.SingleFileData, 0, len(imData))

		for _, val := range imData {
			resp = append(resp, &storagev1.SingleFileData{
				Name:       val.Name,
				UploadTime: &val.UploadTime,
				UpdateTime: val.UpdateTime,
			})
		}

		return &storagev1.FileDataResponse{ImageData: resp}, nil
	default:
		return nil, status.Error(codes.PermissionDenied, "network is busy")
	}
}
