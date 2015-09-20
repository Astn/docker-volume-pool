package main

import (
	"fmt"
	"log"
    "io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"github.com/calavera/dkvolume"
    "github.com/samalba/dockerclient"
)

type volume struct {
	name        string
	connections int
}

type volumePoolDriver struct {
	root       string
	docker	   *dockerclient.DockerClient
	volumes    map[string]*volume
	m          *sync.Mutex
}

// Callback used to listen to Docker's events
func eventCallback(event *dockerclient.Event, ec chan error, args ...interface{}) {
    log.Printf("Received event: %#v\n", *event)
}

func newVolumePoolDriver(root string) volumePoolDriver {
	
	docker, _ := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)

	d := volumePoolDriver{
		root: 	 root,
		docker:  docker,
		volumes: map[string]*volume{},
		m:       &sync.Mutex{},
	}

    // Init the client
    //docker, _ := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)

    // Listen to events
    d.docker.StartMonitorEvents(eventCallback, nil)

	return d
}

func (d volumePoolDriver) Create(r dkvolume.Request) dkvolume.Response {
	log.Printf("Creating volume %s with options %s\n", r.Name, r.Options)
	d.m.Lock()
	defer d.m.Unlock()

    //assign volume to group, defined in r.Options
    group := r.Options["group"]
	log.Printf("assigning to group %s\n", group) 	
	volumeName := fmt.Sprintf("%s-%s", group, r.Name)

    scheme := r.Options["scheme"]
	log.Printf("using scheme %s\n", scheme)

	winner := ""

	files, _ := ioutil.ReadDir(*root)
	
	if true { //scheme == "ordered" {
    	//for scheme=ordered, use directory with least # of group dirs
	    for _, f := range files {

	    	filePath := filepath.Join(*root, f.Name())
		    fmt.Printf("*found %s\n", filePath)

			nestedFiles, _ := ioutil.ReadDir(filePath)
			if len(nestedFiles) == 0 {
				winner = filepath.Join(f.Name())
				break
			} else {
			    for _, nf := range nestedFiles {

			    	nfPath := filepath.Join(*root, f.Name(), nf.Name())
			        fmt.Printf("**found %s\n", nfPath)

			        winner = filepath.Join(f.Name(), nf.Name())
			        break
			    }
			}
		}
	} else {
		//ls for directories in d.root, pick 1st one
		winner = filepath.Join(*root, files[0].Name())
	}

	log.Printf("winner is %s\n", winner)

	mountpoint := filepath.Join(d.root, winner, group, r.Name)

	if _, ok := d.volumes[mountpoint]; ok {
		return dkvolume.Response{}
	}

	log.Printf("creating directory at %s\n", mountpoint)
	err := os.MkdirAll(mountpoint, os.ModeDir)
	if err != nil {
        log.Fatal(err)
    }

	containerName := fmt.Sprintf("pool-volume-%s", volumeName)
	fmt.Sprintf("container name is %s\n", containerName)

 	// Get only running containers
    containers, err := d.docker.ListContainers(true, false, "")
    if err != nil {
        log.Fatal(err)
    }
    
    containerExists := false
    containerIsRunning := false

    containerId := ""

    for _, c := range containers {
        log.Printf("found container %s, %s\n",c.Id, c.Names)
        
        if stringsContain(c.Names, fmt.Sprintf("/%s", containerName)) {
        	containerExists = true
			
			runningContainers, err := d.docker.ListContainers(false, false, "")
		    if err != nil {
		        log.Fatal(err)
		    }
		    for _, rc := range runningContainers {			    
			    if stringsContain(rc.Names, containerName) {
			    	containerIsRunning = true
        		}
        	}
        }	
    }

    if containerExists != true {
    	// Create a container
	    containerConfig := &dockerclient.ContainerConfig{
	        Image: "ubuntu:14.04",
	        Volumes: map[string]struct{}{fmt.Sprintf("%s:/data0", r.Name): {}},
	        Cmd:   []string{"bash"},
	        AttachStdin: true,
	        Tty:   true,
	    }
	    containerId, err = d.docker.CreateContainer(containerConfig, containerName)
	    if err != nil {
	        log.Fatal(err)
	    }
    }

    if containerIsRunning != true {
	    // Start the container
	    hostConfig := &dockerclient.HostConfig{}
	    err = d.docker.StartContainer(containerId, hostConfig)
	    if err != nil {
	        log.Fatal(err)
	    }
    }

	return dkvolume.Response{}
}

func (d volumePoolDriver) Remove(r dkvolume.Request) dkvolume.Response {
	log.Printf("Removing volume %s\n", r.Name)
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)

	if s, ok := d.volumes[m]; ok {
		if s.connections <= 1 {
			
			delete(d.volumes, m)
		}
	}
	return dkvolume.Response{}
}

func (d volumePoolDriver) Path(r dkvolume.Request) dkvolume.Response {
	return dkvolume.Response{Mountpoint: d.mountpoint(r.Name)}
}

func (d volumePoolDriver) Mount(r dkvolume.Request) dkvolume.Response {
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)
	log.Printf("Mounting volume %s on %s\n", r.Name, m)

	s, ok := d.volumes[m]
	if ok && s.connections > 0 {
		s.connections++
		return dkvolume.Response{Mountpoint: m}
	}

	fi, err := os.Lstat(m)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(m, 0755); err != nil {
			return dkvolume.Response{Err: err.Error()}
		}
	} else if err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	if fi != nil && !fi.IsDir() {
		return dkvolume.Response{Err: fmt.Sprintf("%v already exists and is not a directory", m)}
	}

	if err := d.mountVolume(r.Name, m); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	d.volumes[m] = &volume{name: r.Name, connections: 1}

	return dkvolume.Response{Mountpoint: m}
}

func (d volumePoolDriver) Unmount(r dkvolume.Request) dkvolume.Response {
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)
	log.Printf("Unmounting volume %s from %s\n", r.Name, m)

	if s, ok := d.volumes[m]; ok {
		if s.connections == 1 {
			if err := d.unmountVolume(m); err != nil {
				return dkvolume.Response{Err: err.Error()}
			}
		}
		s.connections--
	} else {
		return dkvolume.Response{Err: fmt.Sprintf("Unable to find volume mounted on %s", m)}
	}

	return dkvolume.Response{}
}

func (d *volumePoolDriver) mountpoint(name string) string {
	return filepath.Join(d.root, name)
}

func (d *volumePoolDriver) mountVolume(name, destination string) error {
	//server := d.servers[rand.Intn(len(d.servers))]

	// cmd := fmt.Sprintf("glusterfs --log-level=DEBUG --volfile-id=%s --volfile-server=%s %s", name, server, destination)
	// if out, err := exec.Command("sh", "-c", cmd).CombinedOutput(); err != nil {
	// 	log.Println(string(out))
	// 	return err
	// }
	return nil
}

func (d *volumePoolDriver) unmountVolume(target string) error {
	cmd := fmt.Sprintf("umount %s", target)
	if out, err := exec.Command("sh", "-c", cmd).CombinedOutput(); err != nil {
		log.Println(string(out))
		return err
	}
	return nil
}

func stringsContain(list []string, str string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
