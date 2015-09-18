# docker volume-pool plugin
goal: add support for something similar to docker -P but for volumes. So docker -V would consistantly automount volumes for a given image and container name.

Slacker @ https://dockerdenver.slack.com/

ex:
```
docker run \
  -v 0:/data0 \
  -v 1:/data1 \
  -v 2:/data2 \
  -v 3:/data3 \
  -v 4:/data4 \
  -v 5:/data5 \
  -v 6:/data6 \
  -v 7:/data7 \
  -v 8:/data8 \
  -v 9:/data9 \
  -v 10:/data10 \
  -v 11:/data11 \
  -v 12:/data12 \
  -v 13:/data13 \
  --plugin=volume-pool --volumeroot=/foo \ [--volume-allocation-scheme=(ordered, round-robin, least-allocated, least-iops)]  \
[--volume-no-cache] 
--name elasticsearch \
elasticsearch
```
given a dir at --volume-root
with mounted disks
```
/foo
   /jboddisk0
   /jboddisk1
   /jboddisk2
   /jboddisk3
   /jboddisk4
   /jboddisk5
   /jboddisk6
```
and a --volume-assignment-scheme=ordered
will create on the file system:
```
/foo
   /jboddisk0
      /elasticsearch-elasticsearch-data0
      /elasticsearch-elasticsearch-data7
   /jboddisk1
      /elasticsearch-elasticsearch-data1
      /elasticsearch-elasticsearch-data8
   /jboddisk2
      /elasticsearch-elasticsearch-data2
      /elasticsearch-elasticsearch-data9
   /jboddisk3
      /elasticsearch-elasticsearch-data3
      /elasticsearch-elasticsearch-data10
   /jboddisk4
      /elasticsearch-elasticsearch-data4
      /elasticsearch-elasticsearch-data11
   /jboddisk5
      /elasticsearch-elasticsearch-data5
      /elasticsearch-elasticsearch-data12
   /jboddisk6
      /elasticsearch-elasticsearch-data6
      /elasticsearch-elasticsearch-data13
```
And will then create a docker volume container with volumes for key(imagename:containername)
```
  -v /foo/jboddisk0/elasticsearch-elasticsearch-data0:data0
  -v /foo/jboddisk1/elasticsearch-elasticsearch-data1:data1  
  -v /foo/jboddisk2/elasticsearch-elasticsearch-data1:data2
  -v /foo/jboddisk3/elasticsearch-elasticsearch-data1:data3
  -v /foo/jboddisk4/elasticsearch-elasticsearch-data1:data4
  -v /foo/jboddisk5/elasticsearch-elasticsearch-data1:data5
  -v /foo/jboddisk6/elasticsearch-elasticsearch-data1:data6
  -v /foo/jboddisk0/elasticsearch-elasticsearch-data1:data7
  -v /foo/jboddisk1/elasticsearch-elasticsearch-data1:data8
  -v /foo/jboddisk2/elasticsearch-elasticsearch-data1:data9
  -v /foo/jboddisk3/elasticsearch-elasticsearch-data1:data10
  -v /foo/jboddisk4/elasticsearch-elasticsearch-data1:data11
  -v /foo/jboddisk5/elasticsearch-elasticsearch-data1:data12
  -v /foo/jboddisk6/elasticsearch-elasticsearch-data1:data13
```
and bind that volume container to the elasticsearch container.  

## Build

```
sudo apt-get install golang-go

export GOPATH=~/gocode

git clone <repo-location> ~/gocode/src/<package-name>

cd ~/gocode/<package-name>

go build <package-name>

go install <package-name>

# starts up the plugin and listens to interesting docker events
sudo env "PATH=$PATH" <package-name>
```

now in a separate terminal window use ```docker run -it --rm --volume-driver test_volume -v foo:/bar ubuntu bash``` to create a volume at ```/var/lib/docker-volumes/_pool/foo``` and a volume container called ```pool-volume-foo``` that mounts the volume as ```data0```, which then gets mounted into our container as ```/bar```.

