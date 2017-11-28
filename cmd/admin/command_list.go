package admin

import (
	"context"
	"fmt"
	"io"

	"github.com/chrislusf/vasto/pb"
)

func init() {
	commands = append(commands, &CommandList{})
}

type CommandList struct {
	masterClient pb.VastoMasterClient
}

func (c *CommandList) Name() string {
	return "list"
}

func (c *CommandList) Help() string {
	return "list all nodes in the cluster"
}

func (c *CommandList) SetMasterCilent(masterClient pb.VastoMasterClient) {
	c.masterClient = masterClient
}

func (c *CommandList) Do(args []string, out io.Writer) error {

	dc := "defaultDataCenter"
	if len(args) > 0 {
		dc = args[0]
	}

	listResponse, err := c.masterClient.ListStores(
		context.Background(),
		&pb.ListRequest{
			DataCenter: dc,
		},
	)

	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Cluster Client Count : %d\n", listResponse.ClientCount)
	cluster := listResponse.GetCluster()
	fmt.Fprintf(out, "Cluster Expected Size: %d\n", cluster.ExpectedClusterSize)
	fmt.Fprintf(out, "Cluster Current  Size: %d\n", cluster.CurrentClusterSize)
	if cluster.NextClusterSize != 0 {
		fmt.Fprintf(out, "Cluster is changing to: %d\n", cluster.NextClusterSize)
	}

	for _, node := range cluster.Nodes {
		fmt.Fprintf(out, "%4d: %32v\n", node.GetShardId(), node.GetAddress())
	}

	return nil
}