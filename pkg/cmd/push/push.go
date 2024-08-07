/*
Copyright © 2024 Jozu.com
*/
package push

import (
	"context"
	"fmt"

	"kitops/pkg/lib/repo/local"
	"kitops/pkg/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
)

func PushModel(ctx context.Context, localRepo local.LocalRepo, repo registry.Repository, opts *pushOptions) (ocispec.Descriptor, error) {
	trackedRepo, logger := output.WrapTarget(repo)
	ref := opts.modelRef.Reference
	copyOpts := oras.CopyOptions{}
	copyOpts.Concurrency = opts.Concurrency
	desc, err := oras.Copy(ctx, localRepo, ref, trackedRepo, ref, copyOpts)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to copy to remote: %w", err)
	}
	logger.Wait()

	return desc, err
}
