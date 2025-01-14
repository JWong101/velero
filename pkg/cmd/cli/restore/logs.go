/*
Copyright 2017 the Velero contributors.

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

package restore

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/client"
	"github.com/vmware-tanzu/velero/pkg/cmd"
	"github.com/vmware-tanzu/velero/pkg/cmd/util/downloadrequest"
)

func NewLogsCommand(f client.Factory) *cobra.Command {
	timeout := time.Minute

	c := &cobra.Command{
		Use:   "logs RESTORE",
		Short: "Get restore logs",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			restoreName := args[0]

			veleroClient, err := f.Client()
			cmd.CheckError(err)

			restore, err := veleroClient.VeleroV1().Restores(f.Namespace()).Get(restoreName, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				cmd.Exit("Restore %q does not exist.", restoreName)
			} else if err != nil {
				cmd.Exit("Error checking for restore %q: %v", restoreName, err)
			}

			switch restore.Status.Phase {
			case v1.RestorePhaseCompleted, v1.RestorePhaseFailed, v1.RestorePhasePartiallyFailed:
				// terminal phases, don't exit.
			default:
				cmd.Exit("Logs for restore %q are not available until it's finished processing. Please wait "+
					"until the restore has a phase of Completed or Failed and try again.", restoreName)
			}

			err = downloadrequest.Stream(veleroClient.VeleroV1(), f.Namespace(), restoreName, v1.DownloadTargetKindRestoreLog, os.Stdout, timeout)
			cmd.CheckError(err)
		},
	}

	c.Flags().DurationVar(&timeout, "timeout", timeout, "how long to wait to receive logs")

	return c
}
