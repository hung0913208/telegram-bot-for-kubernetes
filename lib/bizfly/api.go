package bizfly

import (
    "context"
    "time"
    "fmt"

    api "github.com/bizflycloud/gobizfly"
)

type Api interface {
    GetKubeconfig(name string) (string, error)
}

type apiImpl struct {
    ctx        context.Context
    cancelFunc context.CancelFunc
    token      *api.Token
    client     *api.Client
}

func NewApi(
    host string,
    username, password string,
    timeout time.Duration,
) (Api, error) {
    return NewApiWithProjectId(
        host, "",
        username,
        password,
        timeout,
    )
}

func NewApiWithProjectId(
    host string,
    projectId string,
    username, password string,
    timeout time.Duration,
) (Api, error) {
    ctx, cancelFunc := context.WithTimeout(
        context.Background(), 
        time.Millisecond*timeout,
    )
	defer cancelFunc()

    client, err := api.NewClient(api.WithAPIUrl(host))
    if err != nil {
        return nil, err
    }

    token, err := client.Token.Create(
        ctx,
        &api.TokenCreateRequest{
            AuthMethod: "password",
            Username:  username,
            Password:  password,
            ProjectID: projectId,
        },
    )
    if err != nil {
        return nil, err
    }

	client.SetKeystoneToken(token)
    return &apiImpl{
        ctx:        ctx,
        token:      token,
        client:     client,
        cancelFunc: cancelFunc,
    }, nil
}

func (self *apiImpl) GetKubeconfig(name string) (string, error) {
    defer self.cancelFunc()

    // @TODO: caching here and use cache in case we would like to access
    //        cluster information without stressing the RESTful APIs
    clusters, err := self.client.KubernetesEngine.List(self.ctx, nil)
    if err != nil {
        return "", err
    }

    // @TODO: we are store the cluster by name and sort by name so we could
    //        easily find the correctness cluster by name
    for _, cluster := range clusters {
        if name == cluster.Name {
            return self.client.KubernetesEngine.GetKubeConfig(
                self.ctx,
                cluster.UID,
            )
        }
    }

    return "", fmt.Errorf("Not found cluster %s", name)
}

func (self *apiImpl) ListCluster() ([]*api.Cluster, error) { 
    return self.client.KubernetesEngine.List(self.ctx, nil)
}
