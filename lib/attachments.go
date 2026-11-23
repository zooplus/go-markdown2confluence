package lib

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/justmiles/go-confluence"
)

// syncAttachments uploads the local files as attachments of the given content.
//
// An attachment that already exists with the same content is left alone: the
// md5 of the file is stored in the attachment comment, and re-uploading an
// unchanged attachment only adds a version and a chance for the Confluence
// Cloud API to fail (intermittent HTTP 500 UnexpectedRollbackException).
func syncAttachments(client *confluence.Client, contentID string, files []string) (errs []error) {
	for _, file := range files {
		if err := syncAttachment(client, contentID, file); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func syncAttachment(client *confluence.Client, contentID, file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	sum := md5.Sum(data)
	hash := hex.EncodeToString(sum[:])
	filename := filepath.Base(file)

	endpoint := "/rest/api/content/" + contentID + "/child/attachment"
	minorEdit := false

	existing, err := client.GetAttachmentByFilename(contentID, filename)
	if err == nil {
		if existing.Metadata.Comment == hash {
			if client.Debug {
				fmt.Printf("SKIPPING UNCHANGED ATTACHMENT: %s\n", file)
			}
			return nil
		}
		endpoint += "/" + existing.ID + "/data"
		minorEdit = true
	}

	return uploadAttachment(client, endpoint, filename, data, hash, minorEdit)
}

// uploadAttachment POSTs the file, with its md5 as comment, to the attachment endpoint.
func uploadAttachment(client *confluence.Client, endpoint, filename string, data []byte, hash string, minorEdit bool) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err = io.Copy(part, bytes.NewReader(data)); err != nil {
		return err
	}
	if minorEdit {
		if err = writer.WriteField("minorEdit", "true"); err != nil {
			return err
		}
	}
	if err = writer.WriteField("comment", hash); err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, client.Endpoint+endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("X-Atlassian-Token", "no-check")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	switch {
	case client.Cookie != "":
		req.Header.Set("Cookie", fmt.Sprintf("JSESSIONID=%v", client.Cookie))
	case client.AccessToken != "":
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", client.AccessToken))
	default:
		req.SetBasicAuth(client.Username, client.Password)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("uploading attachment %s: %w", filename, err)
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusMultipleChoices {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("uploading attachment %s: HTTP %d: %s", filename, res.StatusCode, msg)
	}
	return nil
}
