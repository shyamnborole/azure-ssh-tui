package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

type Client interface {
	CheckAuth(ctx context.Context) error
	ListSubscriptions(ctx context.Context) ([]Subscription, error)
	ListVMs(ctx context.Context, subscriptionID string) ([]VM, error)
}

type cliClient struct{}

func NewClient() Client {
	return &cliClient{}
}

func (c *cliClient) CheckAuth(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "az", "account", "show", "--output", "json")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("auth check failed: %w, stderr: %s", err, stderr.String())
	}
	return nil
}

func (c *cliClient) ListSubscriptions(ctx context.Context) ([]Subscription, error) {
	cmd := exec.CommandContext(ctx, "az", "account", "list", "--output", "json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w, stderr: %s", err, stderr.String())
	}

	var subs []Subscription
	if err := json.Unmarshal(stdout.Bytes(), &subs); err != nil {
		return nil, fmt.Errorf("failed to parse subscriptions JSON: %w", err)
	}

	return subs, nil
}

func (c *cliClient) ListVMs(ctx context.Context, subscriptionID string) ([]VM, error) {
	args := []string{"vm", "list", "--output", "json"}
	if subscriptionID != "" {
		args = append(args, "--subscription", subscriptionID)
	}

	cmd := exec.CommandContext(ctx, "az", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w, stderr: %s", err, stderr.String())
	}

	var vms []VM
	if err := json.Unmarshal(stdout.Bytes(), &vms); err != nil {
		return nil, fmt.Errorf("failed to parse VMs JSON: %w", err)
	}

	return vms, nil
}
