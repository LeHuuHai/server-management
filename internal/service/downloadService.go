package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type DownloadService struct {
	BaseURL string
	Client  *http.Client
}

func (s *DownloadService) Download(
	ctx context.Context,
	filename string,
) ([]byte, error) {
	url := fmt.Sprintf(
		"%s/%s",
		s.BaseURL,
		filename,
	)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"unexpected status: %d",
			resp.StatusCode,
		)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func NewDownLoadService(base string, client *http.Client) *DownloadService {
	return &DownloadService{
		BaseURL: base,
		Client:  client,
	}
}
