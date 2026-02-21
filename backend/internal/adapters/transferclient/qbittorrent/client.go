package qbittorrent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain/transfer"
	"github.com/sonalys/animeman/internal/utils/roundtripper"
	"google.golang.org/grpc/codes"
)

type (
	Client struct {
		client interface {
			Do(req *http.Request) (*http.Response, error)
		}
	}
)

func New(ctx context.Context, host, username, password string) (*Client, error) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	basePath, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("parsing host address: %w", err)
	}

	apiEndpoint := basePath.JoinPath("api", "v2")

	api := &Client{
		client: &http.Client{
			Jar: cookieJar,
			Transport: roundtripper.NewPrefix(
				apiEndpoint,
				newAuthTransport(
					roundtripper.NewLoggerTransport(http.DefaultTransport),
					username,
					password,
				),
			),
		},
	}

	return api, nil
}

func (c *Client) Wait(ctx context.Context) error {
	var err error

	for {
		if ctx.Err() != nil {
			return errors.Join(err, ctx.Err())
		}

		if _, err = c.Version(ctx); err == nil {
			return nil
		}

		time.Sleep(time.Second)
	}
}

func (c *Client) Version(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http:///app/version", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	textResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", apperr.New(fmt.Errorf("reading version: %s", textResponse), codes.InvalidArgument, "hostname responded unexpectedly")
	}

	return string(textResponse), nil
}

func writePart(w *multipart.Writer, name string, contentType string, buffer []byte) error {
	h := make(textproto.MIMEHeader)
	cd := fmt.Sprintf(`form-data; name="torrents"; filename="%s"`, name)
	h.Set("Content-Disposition", cd)
	h.Set("Content-Type", contentType)

	part, err := w.CreatePart(h)
	if err != nil {
		return err
	}

	if _, err = part.Write(buffer); err != nil {
		return err
	}

	return nil
}

func createReleaseCandidateRequest(ctx context.Context, releaseCandidate transfer.ReleaseCandidate) (*http.Request, error) {
	buffer := bytes.NewBuffer(make([]byte, 1024))
	w := multipart.NewWriter(buffer)

	if err := writePart(w, releaseCandidate.Name, releaseCandidate.ContentType, releaseCandidate.Binary); err != nil {
		return nil, fmt.Errorf("creating 'torrents' part: %w", err)
	}

	field, err := w.CreateFormField("tags")
	if err != nil {
		return nil, fmt.Errorf("creating form field 'tags': %w", err)
	}

	if _, err := io.WriteString(field, strings.Join(releaseCandidate.Tags, ",")); err != nil {
		return nil, fmt.Errorf("writing 'tags' field: %w", err)
	}

	field, err = w.CreateFormField("category")
	if err != nil {
		return nil, fmt.Errorf("creating form field 'category': %w", err)
	}

	if _, err := io.WriteString(field, releaseCandidate.Category); err != nil {
		return nil, fmt.Errorf("writing 'category' field: %w", err)
	}

	field, err = w.CreateFormField("paused")
	if err != nil {
		return nil, fmt.Errorf("creating form field 'paused': %w", err)
	}

	if _, err := io.WriteString(field, fmt.Sprint(releaseCandidate.ShouldPause)); err != nil {
		return nil, fmt.Errorf("writing 'paused' field: %w", err)
	}

	field, err = w.CreateFormField("savepath")
	if err != nil {
		return nil, fmt.Errorf("creating form field 'savepath': %w", err)
	}

	if _, err := io.WriteString(field, fmt.Sprint(releaseCandidate.Filepath)); err != nil {
		return nil, fmt.Errorf("writing 'savepath' field: %w", err)
	}

	field, err = w.CreateFormField("rename")
	if err != nil {
		return nil, fmt.Errorf("creating form field 'rename': %w", err)
	}

	if _, err := io.WriteString(field, releaseCandidate.Name); err != nil {
		return nil, fmt.Errorf("writing 'rename' field: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/torrents/add", buffer)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	return req, nil
}

func (c *Client) Download(ctx context.Context, releaseCandidate transfer.ReleaseCandidate) (*transfer.ReleaseDownload, error) {
	req, err := createReleaseCandidateRequest(ctx, releaseCandidate)
	if err != nil {
		return nil, fmt.Errorf("creating multi-part form: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("creating release candidate: %w", err)
	}

	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("closing response body: %w", err)
	}

	return &transfer.ReleaseDownload{
		ReleaseID:       releaseCandidate.ReleaseID,
		Filepath:        releaseCandidate.Filepath,
		Status:          transfer.StatusPending,
		ProgressDecimal: 0,
	}, nil
}
