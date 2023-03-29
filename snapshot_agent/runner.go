package snapshot_agent

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/wasilak/vault_raft_snapshot_agent/config"
)

func logSnapshotError(dest, snapshotPath string, err error) (string, error) {
	if err != nil {
		return "", fmt.Errorf("failed to generate %s snapshot to %s: %v", dest, snapshotPath, err)
	} else {
		return fmt.Sprintf("successfully created %s snapshot to %s", dest, snapshotPath), nil
	}
}

func RunBackup(snapshotter *Snapshotter, c *config.Configuration) (string, error) {
	if snapshotter.TokenExpiration.Before(time.Now()) {
		switch c.VaultAuthMethod {
		case "k8s":
			snapshotter.SetClientTokenFromK8sAuth(c)
		default:
			snapshotter.SetClientTokenFromAppRole(c)
		}
	}

	leader, err := snapshotter.API.Sys().Leader()
	if err != nil {
		return "", fmt.Errorf("unable to determine leader instance. The snapshot agent will only run on the leader node.  Are you running this daemon on a Vault instance? %s", err)
	}

	leaderIsSelf := leader.IsSelf
	if !leaderIsSelf {
		log.Println("Not running on leader node, skipping.")
	} else {
		var snapshot bytes.Buffer
		err := snapshotter.API.Sys().RaftSnapshot(&snapshot)
		if err != nil {
			return "", fmt.Errorf("unable to generate snapshot, %s", err.Error())
		}
		now := time.Now().UnixNano()
		if c.Local.Path != "" {
			snapshotPath, err := snapshotter.CreateLocalSnapshot(&snapshot, c, now)
			return logSnapshotError("local", snapshotPath, err)
		}
		if c.AWS.Bucket != "" {
			snapshotPath, err := snapshotter.CreateS3Snapshot(&snapshot, c, now)
			return logSnapshotError("aws", snapshotPath, err)
		}
		if c.GCP.Bucket != "" {
			snapshotPath, err := snapshotter.CreateGCPSnapshot(&snapshot, c, now)
			return logSnapshotError("gcp", snapshotPath, err)
		}
		if c.Azure.ContainerName != "" {
			snapshotPath, err := snapshotter.CreateAzureSnapshot(&snapshot, c, now)
			return logSnapshotError("azure", snapshotPath, err)
		}
	}

	return "", nil
}
