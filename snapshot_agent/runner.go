package snapshot_agent

import (
	"bytes"
	"log"
	"time"

	"github.com/wasilak/vault_raft_snapshot_agent/config"
)

func logSnapshotError(dest, snapshotPath string, err error) {
	if err != nil {
		log.Printf("Failed to generate %s snapshot to %s: %v\n", dest, snapshotPath, err)
	} else {
		log.Printf("Successfully created %s snapshot to %s\n", dest, snapshotPath)
	}
}

func RunBackup(snapshotter *Snapshotter, c *config.Configuration) {
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
		log.Println(err.Error())
		log.Fatalln("Unable to determine leader instance.  The snapshot agent will only run on the leader node.  Are you running this daemon on a Vault instance?")
	}
	leaderIsSelf := leader.IsSelf
	if !leaderIsSelf {
		log.Println("Not running on leader node, skipping.")
	} else {
		var snapshot bytes.Buffer
		err := snapshotter.API.Sys().RaftSnapshot(&snapshot)
		if err != nil {
			log.Fatalln("Unable to generate snapshot", err.Error())
		}
		now := time.Now().UnixNano()
		if c.Local.Path != "" {
			snapshotPath, err := snapshotter.CreateLocalSnapshot(&snapshot, c, now)
			logSnapshotError("local", snapshotPath, err)
		}
		if c.AWS.Bucket != "" {
			snapshotPath, err := snapshotter.CreateS3Snapshot(&snapshot, c, now)
			logSnapshotError("aws", snapshotPath, err)
		}
		if c.GCP.Bucket != "" {
			snapshotPath, err := snapshotter.CreateGCPSnapshot(&snapshot, c, now)
			logSnapshotError("gcp", snapshotPath, err)
		}
		if c.Azure.ContainerName != "" {
			snapshotPath, err := snapshotter.CreateAzureSnapshot(&snapshot, c, now)
			logSnapshotError("azure", snapshotPath, err)
		}
	}
}
