/*
Copyright (c) 2020 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"errors"
	"fmt"
	"os"
	"time"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/spf13/cobra"

	"github.com/openshift/rosa/pkg/aws"
	clusterprovider "github.com/openshift/rosa/pkg/cluster"
	"github.com/openshift/rosa/pkg/confirm"
	"github.com/openshift/rosa/pkg/interactive"
	"github.com/openshift/rosa/pkg/logging"
	"github.com/openshift/rosa/pkg/ocm"
	rprtr "github.com/openshift/rosa/pkg/reporter"
)

var args struct {
	clusterKey string

	// Basic options
	expirationTime     string
	expirationDuration time.Duration

	// Networking options
	private bool
}

var Cmd = &cobra.Command{
	Use:   "cluster",
	Short: "Edit cluster",
	Long:  "Edit cluster.",
	Example: `  # Edit a cluster named "mycluster" to make it private
  rosa edit cluster mycluster --private

  # Edit all options interactively
  rosa edit cluster -c mycluster --interactive`,
	Run: run,
}

func init() {
	flags := Cmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(
		&args.clusterKey,
		"cluster",
		"c",
		"",
		"Name or ID of the cluster to edit.",
	)

	// Basic options
	flags.StringVar(
		&args.expirationTime,
		"expiration-time",
		"",
		"Specific time when cluster should expire (RFC3339). Only one of expiration-time / expiration may be used.",
	)
	flags.DurationVar(
		&args.expirationDuration,
		"expiration",
		0,
		"Expire cluster after a relative duration like 2h, 8h, 72h. Only one of expiration-time / expiration may be used.",
	)
	// Cluster expiration is not supported in production
	flags.MarkHidden("expiration-time")
	flags.MarkHidden("expiration")

	// Networking options
	flags.BoolVar(
		&args.private,
		"private",
		false,
		"Restrict master API endpoint to direct, private connectivity.",
	)
}

func run(cmd *cobra.Command, argv []string) {
	reporter := rprtr.CreateReporterOrExit()

	// Check command line arguments:
	clusterKey := args.clusterKey
	if clusterKey == "" {
		if len(argv) != 1 {
			reporter.Errorf(
				"Expected exactly one command line argument or flag containing the name " +
					"or identifier of the cluster",
			)
			os.Exit(1)
		}
		clusterKey = argv[0]
	}

	// Check that the cluster key (name, identifier or external identifier) given by the user
	// is reasonably safe so that there is no risk of SQL injection:
	if !clusterprovider.IsValidClusterKey(clusterKey) {
		reporter.Errorf(
			"Cluster name, identifier or external identifier '%s' isn't valid: it "+
				"must contain only letters, digits, dashes and underscores",
			clusterKey,
		)
		os.Exit(1)
	}

	isInteractive := interactive.Enabled()
	if !isInteractive {
		changedFlags := false
		for _, flag := range []string{"private"} {
			if cmd.Flags().Changed(flag) {
				changedFlags = true
			}
		}
		if !changedFlags {
			isInteractive = true
		}
	}

	logger := logging.CreateLoggerOrExit(reporter)

	// Create the client for the OCM API:
	ocmConnection, err := ocm.NewConnection().
		Logger(logger).
		Build()
	if err != nil {
		reporter.Errorf("Failed to create OCM connection: %v", err)
		os.Exit(1)
	}
	defer func() {
		err = ocmConnection.Close()
		if err != nil {
			reporter.Errorf("Failed to close OCM connection: %v", err)
		}
	}()
	ocmClient := ocmConnection.ClustersMgmt().V1()

	// Create the AWS client:
	awsClient, err := aws.NewClient().
		Logger(logger).
		Build()
	if err != nil {
		reporter.Errorf("Failed to create AWS client: %v", err)
		os.Exit(1)
	}

	awsCreator, err := awsClient.GetCreator()
	if err != nil {
		reporter.Errorf("Failed to get AWS creator: %v", err)
		os.Exit(1)
	}

	reporter.Debugf("Loading cluster '%s'", clusterKey)
	cluster, err := clusterprovider.GetCluster(ocmClient.Clusters(), clusterKey, awsCreator.ARN)
	if err != nil {
		reporter.Errorf("Failed to get cluster '%s': %v", clusterKey, err)
		os.Exit(1)
	}

	// Validate flags:
	expiration, err := validateExpiration()
	if err != nil {
		reporter.Errorf(fmt.Sprintf("%s", err))
		os.Exit(1)
	}

	if interactive.Enabled() {
		reporter.Infof("Interactive mode enabled.\n" +
			"Any optional fields can be ignored and will not be updated.")
	}

	var private *bool
	var privateValue bool
	if cmd.Flags().Changed("private") {
		privateValue = args.private
		private = &privateValue
	} else if isInteractive {
		privateValue = cluster.API().Listening() == cmv1.ListeningMethodInternal
	}

	privateWarning := "You will not be able to access your cluster until you edit network settings in your cloud provider."
	if isInteractive {
		privateValue, err = interactive.GetBool(interactive.Input{
			Question: "Private cluster",
			Help:     fmt.Sprintf("%s %s", cmd.Flags().Lookup("private").Usage, privateWarning),
			Default:  privateValue,
		})
		if err != nil {
			reporter.Errorf("Expected a valid private value: %s", err)
			os.Exit(1)
		}
		private = &privateValue
	} else if privateValue {
		reporter.Warnf("You are choosing to make your cluster private. %s", privateWarning)
		if !confirm.Confirm("set cluster '%s' as private", clusterKey) {
			os.Exit(0)
		}
	}

	clusterConfig := clusterprovider.Spec{
		Expiration: expiration,
		Private:    private,
	}

	reporter.Debugf("Updating cluster '%s'", clusterKey)
	err = clusterprovider.UpdateCluster(ocmClient.Clusters(), clusterKey, awsCreator.ARN, clusterConfig)
	if err != nil {
		reporter.Errorf("Failed to update cluster: %v", err)
		os.Exit(1)
	}
	reporter.Infof("Updated cluster '%s'", clusterKey)
}

func validateExpiration() (expiration time.Time, err error) {
	// Validate options
	if len(args.expirationTime) > 0 && args.expirationDuration != 0 {
		err = errors.New("At most one of 'expiration-time' or 'expiration' may be specified")
		return
	}

	// Parse the expiration options
	if len(args.expirationTime) > 0 {
		t, err := parseRFC3339(args.expirationTime)
		if err != nil {
			err = fmt.Errorf("Failed to parse expiration-time: %s", err)
			return expiration, err
		}

		expiration = t
	}
	if args.expirationDuration != 0 {
		// round up to the nearest second
		expiration = time.Now().Add(args.expirationDuration).Round(time.Second)
	}

	return
}

// parseRFC3339 parses an RFC3339 date in either RFC3339Nano or RFC3339 format.
func parseRFC3339(s string) (time.Time, error) {
	if t, timeErr := time.Parse(time.RFC3339Nano, s); timeErr == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}
