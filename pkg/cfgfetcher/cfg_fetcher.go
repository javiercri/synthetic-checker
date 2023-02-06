package cfgfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/luisdavim/synthetic-checker/pkg/config"
)

type Fetcher struct {
	config []config.Peer
	log    zerolog.Logger
	client *retryablehttp.Client
	sync.RWMutex
}

func New(cfg []config.Peer) (*Fetcher, error) {
	for _, c := range cfg {
		if c.URL == "" {
			return nil, fmt.Errorf("invalid configuration")
		}
	}

	fetcher := &Fetcher{
		config: cfg,
		client: retryablehttp.NewClient(),
		log:    zerolog.New(os.Stderr).With().Timestamp().Str("name", "cfgFetcher").Logger().Level(zerolog.InfoLevel),
	}
	fetcher.client.Logger = &fetcher.log
	return fetcher, nil
}

func (f *Fetcher) AddPeer(p config.Peer) {
	f.Lock()
	defer f.Unlock()
	for _, c := range f.config {
		if c.URL == p.URL {
			return
		}
	}
	f.config = append(f.config, p)
}

func (f *Fetcher) RemovePeer(url string) {
	f.Lock()
	defer f.Unlock()
	for idx, c := range f.config {
		if c.URL == config.TemplatedString(url) {
			f.config = append(f.config[:idx], f.config[idx+1:]...)
			return
		}
	}
}

func (f *Fetcher) GetConfigs() ([]config.Config, error) {
	eg, _ := errgroup.WithContext(context.Background())
	results := make([]config.Config, len(f.config))
	for i, peer := range f.config {
		i, peer := i, peer // https://golang.org/doc/faq#closures_and_goroutines
		eg.Go(func() error {
			res, err := f.get(fmt.Sprintf("%s/config", peer))
			if err == nil {
				results[i] = res
			}
			return err
		})
	}

	return results, eg.Wait()
}

func (f *Fetcher) get(url string) (config.Config, error) {
	var cfg config.Config
	res, err := f.client.Get(url)
	if err != nil {
		return cfg, err
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
