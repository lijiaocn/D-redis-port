//create: 2015/07/15 11:57:52 change: 2015/09/22 11:03:13 author:lijiao
package container

import(
	docker "github.com/fsouza/go-dockerclient"
)

type Handler struct{
	CreateOptions docker.CreateContainerOptions
	RemoveOptions docker.RemoveContainerOptions
	ImageOptions  docker.PullImageOptions
	Auth          docker.AuthConfiguration
	Container     *docker.Container
	Client        *docker.Client
}

func (h *Handler) SetClient(client *docker.Client) {
	h.Client = client
}

func (h *Handler) SetImage(respos, registry, tag string){
	h.ImageOptions = docker.PullImageOptions{
		Repository:     respos    ,
		Registry:       registry  ,
		Tag:            tag       ,
		OutputStream:   nil       ,
		RawJSONStream:  false     ,
	}
	h.CreateOptions.Config.Image=respos+":"+tag
}

func (h *Handler) SetName(name string){
	h.CreateOptions.Name=name
}

func (h *Handler) SetAuth(user, pass, mail, addr string){
	h.Auth = docker.AuthConfiguration{
		Username:       user        ,
		Password:       pass        ,
		Email:          mail        ,
		ServerAddress:  addr        ,
	}
}

func (h *Handler) SetMemLimit(mem, swap int64){
	h.CreateOptions.HostConfig.Memory = mem
	h.CreateOptions.HostConfig.MemorySwap = swap
	h.CreateOptions.Config.Memory = mem
	h.CreateOptions.Config.MemorySwap = swap
}

func (h *Handler) SetNetworkMode(mode string){
	h.CreateOptions.HostConfig.NetworkMode = mode
}

func (h *Handler) SetPortSpec(ports ... string){
	h.CreateOptions.Config.PortSpecs = ports
}

func (h *Handler) SetCmds(cmd ... string){
	h.CreateOptions.Config.Cmd = cmd
}

func (h *Handler) PullImage() error{
	return h.Client.PullImage(h.ImageOptions, h.Auth)
}

func (h *Handler) Create() error{
	var err error
	h.Container, err = h.Client.CreateContainer(h.CreateOptions)
	if err == docker.ErrNoSuchImage{
		err = h.Client.PullImage(h.ImageOptions, h.Auth)
		if err != nil{
			return err
		}
		h.Container, err = h.Client.CreateContainer(h.CreateOptions)
	}
	if err != nil{
		return err
	}
	h.RemoveOptions.ID = h.Container.ID
	return nil
}

func (h *Handler) Remove()error{
	return h.Client.RemoveContainer(h.RemoveOptions)
}

func (h *Handler) Start()error{
	return h.Client.StartContainer(h.Container.ID, h.CreateOptions.HostConfig)
}

func (h *Handler) Inspect()(*docker.Container, error){
	return h.Client.InspectContainer(h.Container.ID)
}

func (h *Handler) IsRunning() (bool, error){
	c, err := h.Client.InspectContainer(h.Container.ID)
	if err != nil{
		return true, err
	}
	return c.State.Running, nil
}

func (h *Handler) ListAll() ([]docker.APIContainers, error){
	cons, err := h.Client.ListContainers(docker.ListContainersOptions{All:true})
	return cons, err
}

func (h *Handler) ListAllName() ([]string, error){
	cons, err := h.Client.ListContainers(docker.ListContainersOptions{All:true})
	if err != nil{
		return nil, err
	}
	names := make([]string, 0)
	for _,c := range cons{
		detail, err := h.Client.InspectContainer(c.ID)
		if err != nil{
			return names, err
		}
		names = append(names, detail.Name)
	}
	return names, nil
}

func DefaultHandler() *Handler {
	config := docker.Config{}
	hostconfig := docker.HostConfig{}

	return &Handler{
				CreateOptions:docker.CreateContainerOptions{
					Name:"",
					Config: &config,
					HostConfig: &hostconfig,
				},
				RemoveOptions:docker.RemoveContainerOptions{
					ID:"",
					RemoveVolumes:true,
					Force:true,
				},
				Container:nil,
			}
}
