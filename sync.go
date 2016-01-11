//create: 2018/09/17 11:54:26 change: 2015/09/22 13:36:10 author:lijiao
package main
import(
	"log"
	"io"
	"sync"
	"fmt"
	"time"
	"strings"
	"errors"
	"os"
	"os/signal"
	"net/http"
	"encoding/json"
	"github.com/lijiaocn/GoPkgs/config"
	"github.com/lijiaocn/GoPkgs/container"
	"github.com/lijiaocn/GoPkgs/generator"
	"github.com/lijiaocn/GoPkgs/terminal"
	docker "github.com/fsouza/go-dockerclient"
)

const(
	Unknown   string = "Unknown"
	Created   string = "Created"
	Running   string = "Running"
	Lost      string = "Lost"
	Startfail string = "Startfail"
)

type ContainerStat struct{
	Stat      string
	Node      string
	AP        string
	Sync      SyncStat
	Obj       interface{}
}

type SyncStat struct{
	RemoteAddr  string    `json:"RemoteAddr"`
	From        string    `json:"From"`
	To          string    `json:"To"`
	Stat        string    `json:"Stat"`
}

type Stats struct{
	RW    sync.RWMutex
	Map   map[string] *ContainerStat   //key is master's address
}

func (s *Stats) SetStat(key string, stat *ContainerStat){
	s.RW.Lock()
	defer s.RW.Unlock()
	s.Map[key] = stat
}

func (s *Stats) UpdateStat(key string, stat string, obj interface{}) bool{
	s.RW.Lock()
	defer s.RW.Unlock()
	c, ok := s.Map[key]
	if ok == true{
		c.Stat = stat
		c.Obj = obj
	}
	return ok
}

func (s *Stats) UpdateSyncStat(key string, syncstat SyncStat) bool{
	s.RW.Lock()
	defer s.RW.Unlock()
	c, ok := s.Map[key]
	if ok == true{
		c.Sync = syncstat
	}
	return ok
}

func (s *Stats) GetStat(key string) (*ContainerStat, bool){
	s.RW.RLock()
	defer s.RW.RUnlock()
	c, ok := s.Map[key]
	return c,ok
}

func printErrors(fails map[string][]error){
	for key,errs := range fails{
		fmt.Printf("%s\n",key)
		for _,err := range errs{
			fmt.Printf("\t%v\n",err)
		}
	}
}

func deleteContainer(handlers map[string] *container.Handler, fails map[string][]error){
		for key, h := range handlers{
			if _,ok := fails[key]; ok == false{
				if err := h.Remove(); err != nil{
					log.Println(err)
				}
			}
		}
}

func createContainer(params RunParams, stats *Stats) (map[string]*container.Handler,map[string][]error){
	nextDockerNode := generator.StrGenerator(params.DockerNodes)
	nextAP := generator.StrGenerator(params.APs)
	containers := make(map[string] *container.Handler)

	fails := make(map[string][]error)
	var mu sync.Mutex
	finished := make(chan interface{}, 10)

	for _,master := range params.SourceMasters{
		node := nextDockerNode()
		ap := nextAP()

		go func(master string){
			var client  *docker.Client
			var err error = nil
			if params.DockerTLS == true{
				client, err = docker.NewTLSClient("tcp://"+node, params.DockerClientCert,
				                       params.DockerClientKey, params.DockerCA)
				if err != nil{
					e := errors.New(fmt.Sprintf("on %s : %v", node, err))
					mu.Lock()
					fails[master] = append(fails[master], e)
					mu.Unlock()
				}
			}else{
				client, err = docker.NewClient("tcp://" + node)
				if err != nil{
					e := errors.New(fmt.Sprintf("on %s : %v", node, err))
					mu.Lock()
					fails[master] = append(fails[master], e)
					mu.Unlock()
				}
			}

			if err == nil{
				handler := container.DefaultHandler()
				handler.SetClient(client)
				handler.SetNetworkMode("host")
				handler.SetImage(params.ImageRepository, params.ImageRegistry, params.ImageTag)
				handler.SetName(strings.Replace(master, ":", "-",-1) + "_to_" +
				                strings.Replace(ap, ":","-", -1))
				handler.SetCmds("sync", "--from=" + master, "--target=" + ap,
								"--auth=" + params.DestCluster,
								"--notify=http://" + params.ServerAddr + "/syncstat")
				err = handler.Create()
				if err != nil{
					e := errors.New(fmt.Sprintf("on %s : %v", node, err))
					mu.Lock()
					fails[master] = append(fails[master], e)
					mu.Unlock()
				}else{
					mu.Lock()
					containers[master] = handler
					mu.Unlock()
					stats.SetStat(master, &ContainerStat{Stat:Created, Node: node, AP: ap,
					              Sync: SyncStat{Stat: Unknown}})
				}
			}

			finished <- 0
		}(master)
	}

	for i := 0; i < len(params.SourceMasters); i++{
		<-finished
	}

	return containers, fails
}

func startContainer(handlers map[string]*container.Handler, stats *Stats){
	finished := make(chan interface{}, 10)
	for key,h := range handlers{
		go func(key string, h *container.Handler){
			if err := h.Start(); err != nil{
				stats.UpdateStat(key, Startfail, err)
				log.Printf("%s %v\n", key, err)
			}else{
				stats.UpdateStat(key, Running, nil)
			}
			finished <- 0
		}(key, h)
	}

	for i := 0; i < len(handlers);i ++{
		<-finished
	}
	return
}

func displayCurStats(masters []string, stats Stats){
	for _, m := range masters{
		stats.RW.RLock()
		s, ok := stats.Map[m]
		stats.RW.RUnlock()
		if ok == true{
			fmt.Printf("%-18s\t%-18s\t%-18s\t%-10s\t%-10s\n",
				       m, s.Node, s.AP, s.Stat, s.Sync.Stat)
		}else{
			fmt.Printf("-18s\t Not Found\n", m)
		}
	}
}

func listen(params RunParams, stats *Stats){
	http.HandleFunc("/syncstat", func(w http.ResponseWriter, r *http.Request){
		if (r.Method != "POST"){
			http.Error(w, "Method Not Allowed", 405)
			return
		}

		if (r.Body == nil){
			http.Error(w, "Payload Empty", 400)
			return
		}

		if(r.ContentLength < 0 || (r.ContentLength == 0 && r.Body != nil)){
			http.Error(w, "Payload Length Unknown", 411)
			return
		}

		if (r.ContentLength > 2048){
			http.Error(w, "Payload Too Large", 413)
			return
		}
	
		buf := make([]byte, r.ContentLength, r.ContentLength)
		_, err := io.ReadFull(r.Body, buf)
		if err != nil {
			log.Printf("%v\n", err)
			http.Error(w, "Request Timeout", 408)
			return
		}
		var syncstat SyncStat
		if err := json.Unmarshal(buf, &syncstat); err != nil{
			log.Printf("%v\n", err)
			http.Error(w, "Payload Format Wrong", 400)
			return
		}

		syncstat.RemoteAddr = r.RemoteAddr
		stats.UpdateSyncStat(syncstat.From, syncstat)

		return
	})
	log.Fatal(http.ListenAndServe(params.ServerAddr, nil))
}

func clean(exit chan os.Signal, handlers map[string]*container.Handler, stops []chan interface{}){
	<-exit
	for _, stop := range stops{
		stop <- 0
	}
	for key, h := range handlers{
		if err := h.Remove(); err != nil{
			log.Println(key, err)
		}
	}
	os.Exit(0)
}

func monitor(handlers map[string]*container.Handler, stats *Stats, stop chan interface{}){
	for true{
		select {
		case <-stop:
			return
		case <-time.After(time.Second):
			for key,h := range handlers{
				r, err := h.IsRunning()
				if err != nil{
					log.Println(err)
				}
				if r == false{
					stats.UpdateStat(key, Lost, err)
				}
			}
		}
	}
}

func startSync(args map[string]interface{}){
	var params RunParams
	cfile, _ := args["--config"].(string)
	if err := config.LoadConfig(cfile, &params); err != nil{
		log.Fatal(err)
	}
	verify(&params)
	stops := make([]chan interface{},0,0)

	stopmonitor := make(chan interface{}, 1)
	stops = append(stops, stopmonitor)

	stats := Stats{Map:make(map[string]*ContainerStat)}
	go listen(params, &stats)


	handlers, fails := createContainer(params, &stats)
	if len(fails) != 0{
		deleteContainer(handlers, fails)
		printErrors(fails)
		log.Fatal("Create Container Fail!")
	}
	
	exitChan := make(chan os.Signal, 20)
	signal.Notify(exitChan, os.Kill, os.Interrupt)
	go clean(exitChan, handlers, stops)

	startContainer(handlers, &stats)
	go monitor(handlers, &stats, stopmonitor)

	for true{
		terminal.Reset()
		fmt.Printf("%s\n", time.Now())
		fmt.Printf("%-18s\t%-18s\t%-18s\t%-10s\t%-10s\n",
		           "master", "node", "ap", "stat", "sync")
		displayCurStats(params.SourceMasters, stats)
		time.Sleep(3 * time.Second)
	}
}
