package lib

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/justmiles/go-confluence"
)

const emptyInputMD5 = "d41d8cd98f00b204e9800998ecf8427e"

// fakeConfluence records the attachment uploads it receives and answers the
// attachment lookup with the configured stored comment.
type fakeConfluence struct {
	mu            sync.Mutex
	exists        bool
	storedComment string
	failUploads   bool
	uploads       []upload
}

type upload struct {
	path, comment, minorEdit string
}

func (f *fakeConfluence) handler(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		if !f.exists {
			_, _ = w.Write([]byte(`{"results":[],"size":0}`))
			return
		}
		_, _ = w.Write([]byte(`{"results":[{"id":"att1","title":"pic.png","metadata":{"comment":"` + f.storedComment + `"}}],"size":1}`))
		return
	}

	_ = r.ParseMultipartForm(1 << 20)
	f.uploads = append(f.uploads, upload{r.URL.Path, r.FormValue("comment"), r.FormValue("minorEdit")})
	if f.failUploads {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"statusCode":500,"message":"UnexpectedRollbackException"}`))
		return
	}
	_, _ = w.Write([]byte(`{"id":"att1","results":[{"id":"att1"}]}`))
}

func newFixture(t *testing.T, f *fakeConfluence) (*confluence.Client, string, string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(f.handler))
	t.Cleanup(srv.Close)

	content := "image bytes"
	file := filepath.Join(t.TempDir(), "pic.png")
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	sum := md5.Sum([]byte(content))
	return &confluence.Client{Endpoint: srv.URL, Username: "u", Password: "p"}, file, hex.EncodeToString(sum[:])
}

func TestUnchangedAttachmentIsNotReuploaded(t *testing.T) {
	f := &fakeConfluence{exists: true}
	client, file, fileMD5 := newFixture(t, f)
	f.storedComment = fileMD5

	errs := syncAttachments(client, "123", []string{file})

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(f.uploads) != 0 {
		t.Fatalf("expected no upload for an unchanged attachment, got %v", f.uploads)
	}
}

func TestChangedAttachmentIsUpdatedWithCorrectHash(t *testing.T) {
	// the hash written by older versions of this tool was always the md5 of empty input
	f := &fakeConfluence{exists: true, storedComment: emptyInputMD5}
	client, file, fileMD5 := newFixture(t, f)

	errs := syncAttachments(client, "123", []string{file})

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(f.uploads) != 1 || f.uploads[0].path != "/rest/api/content/123/child/attachment/att1/data" {
		t.Fatalf("expected exactly one update upload, got %v", f.uploads)
	}
	if got := f.uploads[0].comment; got != fileMD5 {
		t.Fatalf("stored hash = %q, want the md5 of the file content %q", got, fileMD5)
	}
	if f.uploads[0].minorEdit != "true" {
		t.Fatalf("expected a minor edit, got %q", f.uploads[0].minorEdit)
	}
}

func TestNewAttachmentIsAddedWithCorrectHash(t *testing.T) {
	f := &fakeConfluence{exists: false}
	client, file, fileMD5 := newFixture(t, f)

	errs := syncAttachments(client, "123", []string{file})

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(f.uploads) != 1 || f.uploads[0].path != "/rest/api/content/123/child/attachment" {
		t.Fatalf("expected exactly one add upload, got %v", f.uploads)
	}
	if got := f.uploads[0].comment; got != fileMD5 {
		t.Fatalf("stored hash = %q, want the md5 of the file content %q", got, fileMD5)
	}
}

func TestFailedUploadIsReported(t *testing.T) {
	f := &fakeConfluence{exists: true, storedComment: emptyInputMD5, failUploads: true}
	client, file, _ := newFixture(t, f)

	errs := syncAttachments(client, "123", []string{file})

	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "HTTP 500") {
		t.Fatalf("expected one HTTP 500 error, got %v", errs)
	}
}

func TestMissingLocalFileIsReported(t *testing.T) {
	f := &fakeConfluence{exists: false}
	client, _, _ := newFixture(t, f)

	errs := syncAttachments(client, "123", []string{filepath.Join(t.TempDir(), "missing.png")})

	if len(errs) != 1 || len(f.uploads) != 0 {
		t.Fatalf("expected one error and no upload, got errs=%v uploads=%v", errs, f.uploads)
	}
}
