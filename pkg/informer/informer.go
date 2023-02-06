package informer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/luisdavim/synthetic-checker/pkg/api"
	"github.com/luisdavim/synthetic-checker/pkg/config"
)

// Informer allows syncing check configuration to upstream synthetic-checkers
type Informer struct {
	config []config.Peer
	log    zerolog.Logger
	client *retryablehttp.Client
	sync.RWMutex
}

// New creates a new Informer
func New(config []config.Peer) (*Informer, error) {
	for _, c := range config {
		if c.URL == "" {
			return nil, fmt.Errorf("invalid configuration")
		}
	}

	informer := &Informer{
		config: config,
		client: retryablehttp.NewClient(),
		log:    zerolog.New(os.Stderr).With().Timestamp().Str("name", "informer").Logger().Level(zerolog.InfoLevel),
	}

	informer.client.Logger = &informer.log
	return informer, nil
}

func (i *Informer) AddUpstream(u config.Peer) {
	i.Lock()
	defer i.Unlock()
	for _, c := range i.config {
		if c.URL == u.URL {
			return
		}
	}
	i.config = append(i.config, u)
}

func (i *Informer) RemoveUpstream(url string) {
	i.Lock()
	defer i.Unlock()
	for idx, c := range i.config {
		if c.URL == config.TemplatedString(url) {
			i.config = append(i.config[:idx], i.config[idx+1:]...)
			return
		}
	}
}

// CreateOrUpdate sends the given check configuration to the configured upstreams
func (i *Informer) CreateOrUpdate(check api.Check) error {
	t, n, c, err := check.Config()
	if err != nil {
		return err
	}
	return i.informUpstreams(context.Background(), http.MethodPost, fmt.Sprintf("checks/%s/%s", t, n), c)
}

// Delete deletes the given check configuration from the configured upstreams
func (i *Informer) Delete(check api.Check) error {
	t, n, _, err := check.Config()
	if err != nil {
		return err
	}
	return i.informUpstreams(context.Background(), http.MethodDelete, fmt.Sprintf("checks/%s/%s", t, n), "")
}

func (i *Informer) Replace(check api.Check) error {
	t, n, c, err := check.Config()
	if err != nil {
		return err
	}
	errD := i.informUpstreams(context.Background(), http.MethodDelete, fmt.Sprintf("checks/%s/%s", t, n), "")
	errU := i.informUpstreams(context.Background(), http.MethodPost, fmt.Sprintf("checks/%s/%s", t, n), c)

	if errD != nil || errU != nil {
		return fmt.Errorf("delete err: %v; update err: %v", errD, errU)
	}

	return nil
}

// DeleteByName removes the given check configuration from the configured upstreams
func (i *Informer) DeleteByName(name string) error {
	return i.informUpstreams(context.Background(), http.MethodDelete, fmt.Sprintf("checks/%s", name), "")
}

func (i *Informer) informUpstreams(ctx context.Context, method, endpoint, body string) error {
	eg, _ := errgroup.WithContext(context.Background())
	i.RLock()
	for _, c := range i.config {
		cfg := c
		eg.Go(func() error {
			url := fmt.Sprintf("%s/%s", cfg.URL, endpoint)
			err := i.inform(ctx, cfg.Headers, method, url, body)
			i.log.Log().Err(err).Msgf("informing %q of %s", url, method)
			return err
		})
	}
	i.RUnlock()

	return eg.Wait()
}

func (i *Informer) inform(ctx context.Context, headers map[string]config.TemplatedString, method, url, body string) error {
	req, err := retryablehttp.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request for %q: %w", url, err)
	}

	for h, v := range headers {
		req.Header.Add(h, string(v))
	}

	return i.do(req)
}

func (i *Informer) do(req *retryablehttp.Request) error {
	resp, err := i.client.Do(req)
	if err != nil {
		if b, e := io.ReadAll(req.Body); e == nil {
			if len(b) != 0 {
				err = fmt.Errorf("%w: %s", err, string(b))
			}
		} else {
			err = fmt.Errorf("%w: and failed to read response: %v", err, e)
		}
		return err
	}

	_ = resp.Body.Close()

	return nil
}
