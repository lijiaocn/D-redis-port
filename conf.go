//create: 2015/09/17 11:57:27 change: 2016/01/11 16:34:48 author:lijiao
package main
import(
	"log"
	"fmt"
	"errors"
	"github.com/lijiaocn/GoPkgs/config"
	"github.com/garyburd/redigo/redis"
	docker "github.com/fsouza/go-dockerclient"
)

type RunParams struct{
	SourceMasters []string `json:"SourceMasters"`
	DestCluster   string   `json:"DestCluster"`
	APs           []string `json:"APs"`
	ImageRepository    string `json:"ImageRepository""`
	ImageRegistry      string `json:"ImageRegistry"`
	ImageTag           string `json:"ImageTag"`
	DockerTLS           bool     `json:"DockerTLS"`
	DockerClientCert    string   `json:"DockerClientCert"`
	DockerClientKey     string   `json:"DockerClientKey"`
	DockerCA            string   `json:"DockerCA"`
	DockerNodes         []string `json:"DockerNodes"`
	ServerAddr          string   `json:"ServerAddr"`
}

func checkRedis(addr string) error{
	c, err := redis.Dial("tcp", addr)
	if err != nil{
		log.Printf("Connect Redis Fail: %s  %v\n", addr, err)
		return err
	}
	defer c.Close()
	return nil
}

func checkMaster(addr string) error{
	return checkRedis(addr)
}

func checkAP(addr string) error{
	return checkRedis(addr)
}

func checkDocker(addr string, tls bool, cert, key, ca string ) error{
	if tls == false{
		client, err := docker.NewClient("tcp://"+addr)
		if  err != nil{
			log.Printf("Connect Docker Fail: %s %v\n", addr, err)
			return err
		}
		if err = client.Ping(); err != nil{
			log.Printf("Connect Docker Fail: %s %v\n", addr, err)
			return err
		}
	}
	if tls == true{
		client, err := docker.NewTLSClient("tcp://"+addr, cert, key, ca)
		if err != nil{
			log.Printf("Connect Docker Fail: %s %v\n", addr, err)
			return err
		}
		if err = client.Ping(); err != nil{
			log.Printf("Connect Docker Fail: %s %v\n", addr, err)
			return err
		}
	}
	return nil
}


func checkDestCluster(ap, cluster string) error{
	c, err := redis.Dial("tcp", ap)
	defer c.Close()
	if err != nil{
		log.Printf("Connect Fail: %s  %v\n", ap, err)
		return err
	}
	
	if _, err := c.Do("auth", cluster); err != nil{
		log.Printf("Auth Fail: %s  %s  %v\n", ap, cluster, err)
		return err
	}

	return nil
}

func verifyIntegrity(params *RunParams) error{
	err_incomplete := errors.New("Config file is incomplete")
	var err error = nil
	if len(params.SourceMasters) == 0 {
		log.Printf("SourceMasters is empty")
		err = err_incomplete
	}

	if len(params.APs) == 0 {
		log.Printf("APs is empty")
		err = err_incomplete
	}

	if len(params.ImageRepository) == 0 {
		log.Printf("ImageRepository is empty")
		err = err_incomplete
	}

	if len(params.ImageRegistry) == 0 {
		log.Printf("ImageRegistry is empty")
		err = err_incomplete
	}

	if len(params.ImageTag) == 0 {
		log.Printf("ImageTag is empty")
		err = err_incomplete
	}

	if params.DockerTLS == true {
		if len(params.DockerClientCert) == 0 {
			log.Printf("DockerClientCert is empty")
			err = err_incomplete
		}
		if len(params.DockerClientKey) == 0 {
			log.Printf("DockerClientKey is empty")
			err = err_incomplete
		}
		if len(params.DockerCA) == 0 {
			log.Printf("DockerCA is empty")
			err = err_incomplete
		}
	}

	if len(params.DockerNodes) == 0 {
		log.Printf("DockerNodes is empty")
		err = err_incomplete
	}

	if len(params.ServerAddr) == 0 {
		log.Printf("ServerAddr is empty")
		err = err_incomplete
	}
	return err
}

func verifyAvailability(params *RunParams) error{
	var err error = nil
	for _,master := range params.SourceMasters{
		err = checkMaster(master)
	}

	for _,ap := range params.APs{
		err = checkAP(ap)
		err = checkDestCluster(ap, params.DestCluster)
	}

	for _,node := range params.DockerNodes{
		err = checkDocker(node, params.DockerTLS, params.DockerClientCert, params.DockerClientKey,
		                    params.DockerCA)
	}

	if err != nil{
		err = errors.New("Some services are not available")
	}
	return err
}

func verify(params *RunParams){
	err := verifyIntegrity(params)
	if err != nil{
		log.Fatal(err)
	}

	err = verifyAvailability(params)
	if err != nil{
		log.Fatal(err)
	}
}

func conf(args map[string]interface{}){
	var params RunParams
	params.SourceMasters = make([]string, 0,0)
	params.SourceMasters = append(params.SourceMasters, "192.168.1.1:6379", "192.168.1.2:7000")
	params.DestCluster = "/redis/cluster/21:1417633162239"
	params.APs = make([]string, 0, 0)
	params.APs = append(params.APs, "192.168.192.39:5360", "192.168.192.38:5360")
	params.ImageRepository = "registryURL/redis-port"
	params.ImageRegistry  = "registryURL"
	params.ImageTag  = "latest"
	params.DockerTLS = false
	params.DockerClientCert = "/PATH/client.crt"
	params.DockerClientKey = "/PATH/client.key"
	params.DockerCA = "/PATH/ca.crt"
	params.DockerNodes = make([]string, 0, 0)
	params.DockerNodes = append(params.DockerNodes, "192.168.192.10:2379", "192.168.192.11:2380")
	params.ServerAddr = "127.0.0.1:80"
	tmpfile,_ := args["--template"].(string)
	if bytes,err := config.GenConfig(params, tmpfile); err!=nil{
		log.Fatal(err)
	}else{
		fmt.Printf("%s\n",string(bytes))
	}
}
