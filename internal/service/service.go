package service

type Services interface {
	FileStorage
}

type Service struct {
	FileStorage FileStorage
}

func NewServices(storagePath string) *Service {
	return &Service{
		FileStorage: NewFileStorageService(storagePath),
	}
}
