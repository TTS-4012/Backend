package judge

//
//import (
//	"context"
//	"github.com/docker/docker/api/types"
//	"github.com/docker/docker/api/types/container"
//	docker "github.com/docker/docker/client"
//	"ocontest/pkg/configs"
//
//	//"github.com/docker/docker/daemon/network"
//
//	"github.com/docker/go-connections/nat"
//)
//
//type RunnerOrchestrator interface {
//}
//
//type RunnerOrchestratorImp struct {
//	dockerClient *docker.Client
//}
//
//func NewRunnerOrchestrator(config configs.SectionRunner) (RunnerOrchestrator, error) {
//	dockerClient, err := docker.NewClientWithOpts()
//	if err != nil {
//		return nil, err
//	}
//	return &RunnerOrchestratorImp{dockerClient: dockerClient}, nil
//}
//func (r RunnerOrchestratorImp) CreateNetworks(ctx context.Context, prefix string) error{
//	// create bind network 'local'
//	// create bridge network 'judge'
//	_, err := r.dockerClient.NetworkCreate(ctx, prefix+"_runners_no_internet", types.NetworkCreate{
//		Driver:   "bridge",
//		Internal: true,
//	})
//	return err
//
//}
//func RunContainer(image string) (containerID string, err error) {
//	cli, err := client.NewClientWithOpts()
//	if err != nil {
//		panic(err)
//	}
//
//	ctx := context.Background()
//	resp, err := cli.ContainerCreate(ctx, &container.Config{
//		Image: image,
//	}, &container.HostConfig{
//		PortBindings: map[nat.Port][]nat.PortBinding{nat.Port("8080"): {{HostIP: "127.0.0.1", HostPort: "8080"}}},
//	}, nil, nil, "")
//	if err != nil {
//		panic(err)
//	}
//
//	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
//		panic(err)
//	}
//
//}
