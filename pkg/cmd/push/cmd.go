// Copyright 2024 The KitOps Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package push

import (
	"context"
	"fmt"

	"kitops/pkg/cmd/options"
	"kitops/pkg/lib/constants"
	"kitops/pkg/lib/repo/local"
	"kitops/pkg/lib/repo/remote"
	"kitops/pkg/lib/repo/util"
	"kitops/pkg/output"

	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const (
	shortDesc = `Upload a modelkit to a specified registry`
	longDesc  = `This command pushes modelkits to a remote registry.

The modelkits should be tagged with the target registry and repository before
they can be pushed`

	example = `# Push the latest modelkits to a remote registry
kit push registry.example.com/my-model:latest

# Push a specific version of a modelkits using a tag:
kit push registry.example.com/my-model:1.0.0`
)

type pushOptions struct {
	options.NetworkOptions
	configHome string
	modelRef   *registry.Reference
}

func (opts *pushOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome

	modelRef, extraTags, err := util.ParseReference(args[0])
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}
	if modelRef.Registry == "localhost" {
		return fmt.Errorf("registry is required when pushing")
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("reference cannot include multiple tags")
	}
	opts.modelRef = modelRef

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	return nil
}

func PushCommand() *cobra.Command {
	opts := &pushOptions{}
	cmd := &cobra.Command{
		Use:     "push [flags] registry/repository[:tag|@digest]",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(opts),
	}

	cmd.Args = cobra.ExactArgs(1)
	opts.AddNetworkFlags(cmd)
	cmd.Flags().SortFlags = false

	return cmd
}

func runCommand(opts *pushOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		remoteRepo, err := remote.NewRepository(
			cmd.Context(),
			opts.modelRef.Registry,
			opts.modelRef.Repository,
			&opts.NetworkOptions,
		)
		if err != nil {
			return output.Fatalln(err)
		}

		localRepo, err := local.NewLocalRepo(constants.StoragePath(opts.configHome), opts.modelRef)
		if err != nil {
			return output.Fatalln(err)
		}

		output.Infof("Pushing %s", opts.modelRef.String())
		desc, err := PushModel(cmd.Context(), localRepo, remoteRepo, opts)
		if err != nil {
			return output.Fatalf("Failed to push: %s. Check you have write permissions to the organization and repository you are pushing to.", err)
		}
		output.Infof("Pushed %s", desc.Digest)
		return nil
	}
}
