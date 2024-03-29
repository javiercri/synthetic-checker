package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
	konfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/luisdavim/synthetic-checker/pkg/api"
	"github.com/luisdavim/synthetic-checker/pkg/config"
)

var _ api.Check = &k8sCheck{}

type k8sCheck struct {
	name   string
	config *config.K8sCheck
	client client.Reader
}

var k8sClient client.Reader

// NewK8sCheck returns a Kubernetes resource status check for the given configuration
// The status of a resource is determined by inspecting its status conditions
func NewK8sCheck(name string, config config.K8sCheck) (api.Check, error) {
	if name == "" {
		return nil, fmt.Errorf("CheckName must not be empty")
	}
	if config.Interval.Duration == 0 {
		config.Interval = metav1.Duration{Duration: 30 * time.Second}
	}
	if config.Timeout.Duration == 0 {
		config.Timeout = metav1.Duration{Duration: time.Second}
	}

	if k8sClient == nil {
		kfg, err := konfig.GetConfig()
		if err != nil {
			return nil, err
		}
		if c, err := client.New(kfg, client.Options{}); err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		} else {
			k8sClient = c
		}
	}

	return &k8sCheck{
		name:   name,
		config: &config,
		client: k8sClient,
	}, nil
}

func (c *k8sCheck) Equal(other *k8sCheck) bool {
	return c.config.Equal(*other.config)
}

func (c *k8sCheck) Config() (string, string, string, error) {
	b, err := json.Marshal(c.config)
	if err != nil {
		return "", "", "", err
	}
	return "k8s", c.name, string(b), nil
}

// Interval indicates how often the check should be performed
func (c *k8sCheck) Interval() metav1.Duration {
	return c.config.Interval
}

// InitialDelay indicates how long to delay the check start
func (c *k8sCheck) InitialDelay() metav1.Duration {
	return c.config.InitialDelay
}

// Interval indicates how often the check should be performed
func (c *k8sCheck) Execute(ctx context.Context) (bool, error) {
	ul, err := c.do(ctx)
	if err != nil {
		return false, err
	}

	resCount := len(ul.Items)
	if resCount == 0 {
		return false, fmt.Errorf("no resources found")
	}

	if resCount < c.config.MinCount {
		return false, fmt.Errorf("not enough resources found, expected %d, got %d", c.config.MinCount, resCount)
	}

	allOK := true
	var errs []error
	for _, u := range ul.Items {
		res, err := status.Compute(&u)
		if err != nil {
			allOK = false
			errs = append(errs, err)
			continue
		}
		if ok := res.Status == status.CurrentStatus; !ok {
			allOK = false
			errs = append(errs, fmt.Errorf("%s: wrong resource state: %s - %s", u.GetName(), res.Status, res.Message))
		}
	}

	errCount := len(errs)
	for _, e := range errs {
		err = fmt.Errorf("%w;", e)
	}

	if err != nil {
		err = fmt.Errorf("%d of %d resources are not ok: %w", errCount, resCount, err)
	}
	return allOK, err
}

func (c *k8sCheck) do(ctx context.Context) (*unstructured.UnstructuredList, error) {
	ul := &unstructured.UnstructuredList{}
	gvk, gk := schema.ParseKindArg(c.config.Kind)
	if gvk == nil {
		// this looks strange but it should make sense if you read the ParseKindArg docs
		gvk = &schema.GroupVersionKind{
			Kind:    gk.Kind,
			Version: gk.Group,
		}
	}
	ul.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind + "List", // TODO: is there a better way?
	})

	if c.config.Name != "" {
		// fetching a single resource by name
		u := unstructured.Unstructured{}
		u.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   gvk.Group,
			Kind:    gvk.Kind,
			Version: gvk.Version,
		})
		if err := c.client.Get(ctx, client.ObjectKey{
			Namespace: c.config.Namespace,
			Name:      c.config.Name,
		}, &u); err != nil {
			return nil, fmt.Errorf("failed to get: %w", err)
		}
		ul.Items = append(ul.Items, u)
		return ul, nil
	}

	opts := &client.ListOptions{
		Namespace: c.config.Namespace,
	}
	if c.config.LabelSelector != "" {
		var err error
		opts.LabelSelector, err = labels.Parse(c.config.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("invalid label selector: %w", err)
		}
	}
	if c.config.FieldSelector != "" {
		var err error
		opts.FieldSelector, err = fields.ParseSelector(c.config.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("invalid field selector: %w", err)
		}
	}
	if err := c.client.List(ctx, ul, opts); err != nil {
		return nil, fmt.Errorf("failed to list: %w", err)
	}

	return ul, nil
}
