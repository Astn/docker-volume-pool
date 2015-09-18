package main

import (
	"fmt"
	"log"
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
	log.Printf("Creating volume %s\n", r.Name)
	d.m.Lock()
	defer d.m.Unlock()
	m := d.mountpoint(r.Name)

	if _, ok := d.volumes[m]; ok {
		return dkvolume.Response{}
	}

	containerName := fmt.Sprintf("pool-volume-%s", r.Name)

	// Create a container
    containerConfig := &dockerclient.ContainerConfig{
        Image: "ubuntu:14.04",
        Volumes: map[string]struct{}{fmt.Sprintf("%s:/data0", m): {}},
        Cmd:   []string{"bash"},
        AttachStdin: true,
        Tty:   true}
    containerId, err := d.docker.CreateContainer(containerConfig, containerName)
    if err != nil {
        log.Fatal(err)
    }

    // Start the container
    hostConfig := &dockerclient.HostConfig{}
    err = d.docker.StartContainer(containerId, hostConfig)
    if err != nil {
        log.Fatal(err)
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