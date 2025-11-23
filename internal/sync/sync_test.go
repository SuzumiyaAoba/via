package sync_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/SuzumiyaAoba/entry/internal/sync"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sync Suite")
}

var _ = Describe("Client", func() {
	var (
		server *httptest.Server
		client *sync.Client
		origURL string
	)

	BeforeEach(func() {
		origURL = sync.GitHubAPIURL
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Path == "/gists/gist123" && r.Method == "GET" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"files": {
						"config.yml": {
							"content": "version: \"1\"\nrules: []"
						}
					}
				}`))
				return
			}
			if r.URL.Path == "/gists/gist123" && r.Method == "PATCH" {
				w.WriteHeader(http.StatusOK)
				return
			}
			if r.URL.Path == "/gists" && r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"id": "newgist123"}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		sync.GitHubAPIURL = server.URL
		client = sync.NewClient("token")
	})

	AfterEach(func() {
		server.Close()
		sync.GitHubAPIURL = origURL
	})

	It("should get gist", func() {
		cfg, err := client.GetGist("gist123")
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Version).To(Equal("1"))
	})

	It("should update gist", func() {
		cfg := &config.Config{Version: "1"}
		err := client.UpdateGist("gist123", cfg)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should create gist", func() {
		cfg := &config.Config{Version: "1"}
		id, err := client.CreateGist(cfg, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(id).To(Equal("newgist123"))
	})

	It("should return error on 404", func() {
		_, err := client.GetGist("unknown")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to get gist"))
	})
})
